package whatsapp

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"encore.app/imolink"
	"encore.app/internal/pkg/openaicli"
	"encore.app/internal/pkg/trello"
	"encore.app/session"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	walog "go.mau.fi/whatsmeow/util/log"
)

const (
	assistantInitTimeout = 30 * time.Second
	messageTimeout       = 120 * time.Second
)

var (
	db = sqldb.NewDatabase("whatsapp", sqldb.DatabaseConfig{
		Migrations: "./migrations",
	})

	secrets struct {
		OpenAIKey    string
		TrelloAPIKey string
		TrelloSecret string
	}
)

// Service is the main service for the WhatsApp API.
//
//encore:service
type Service struct {
	whatsappCli *whatsmeow.Client
	deviceStore *store.Device
	clientLock  sync.Mutex
	sessionMgr  *session.SessionManager
	openAICli   *openaicli.Client
	trelloAPI   *trello.TrelloAPI
}

func initService() (*Service, error) {
	s := new(Service)

	if err := imolink.InitializeAssistant(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize assistant: %w", err)
	}

	s.trelloAPI = trello.NewTrelloAPI(secrets.TrelloAPIKey, secrets.TrelloSecret)

	s.openAICli = openaicli.New(
		secrets.OpenAIKey,
		&http.Client{
			Timeout: assistantInitTimeout,
		},
	)

	s.sessionMgr = session.NewSessionManager(imolink.Assistant, s.openAICli)

	dbLog := walog.Stdout("whatsapp-database", "INFO", true)
	container := sqlstore.NewWithDB(db.Stdlib(), "postgres", dbLog)

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		if strings.Contains(err.Error(), "no devices found") {
			return s, nil
		}
		return nil, fmt.Errorf("could not get WhatsApp device: %w", err)
	}

	if deviceStore != nil {
		if err := s.connectToWhatsApp(deviceStore); err != nil {
			return nil, fmt.Errorf("could not connect to WhatsApp: %w", err)
		}
	}
	s.deviceStore = deviceStore
	return s, nil
}

//encore:api public raw path=/whatsapp/connect
func (s *Service) WhatsappConnect(w http.ResponseWriter, req *http.Request) {
	s.clientLock.Lock()
	defer s.clientLock.Unlock()

	// If we have a client, check if it's really connected
	if s.whatsappCli != nil {
		if s.whatsappCli.IsConnected() {
			if s.whatsappCli.IsLoggedIn() {
				fmt.Fprintf(w, "Already connected and logged in to WhatsApp")
				return
			}
			// If connected but not logged in, disconnect and reconnect
			s.whatsappCli.Disconnect()
		}
	}

	clientLog := walog.Stdout("Client", "DEBUG", true)
	container := sqlstore.NewWithDB(db.Stdlib(), "postgres", clientLog)
	deviceStore := container.NewDevice()

	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(s.whatsappEventHandler)

	s.whatsappCli = client
	s.deviceStore = deviceStore

	qrChan, err := client.GetQRChannel(context.Background())
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get QR channel: %v", err), http.StatusInternalServerError)
		return
	}

	if err := client.Connect(); err != nil {
		http.Error(w, fmt.Sprintf("failed to connect: %v", err), http.StatusInternalServerError)
		return
	}

	for evt := range qrChan {
		if evt.Event == "code" {
			fmt.Fprintf(w, "Scan this QR code with WhatsApp:\n\n")
			qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, w)
			return
		} else {
			fmt.Fprintf(w, "Login event: %s\n", evt.Event)
		}
	}

	if client.IsLoggedIn() {
		fmt.Fprintf(w, "Successfully connected to WhatsApp")
	} else {
		fmt.Fprintf(w, "Failed to connect to WhatsApp")
	}
}

func (s *Service) whatsappEventHandler(evt any) {
	switch v := evt.(type) {
	case *events.Message:
		rlog.Debug(
			"Message received",
			"message", v.Message.GetConversation(),
			"sender", v.Info.Sender,
			"target", v.Info.Sender.User,
		)

		if s.sessionMgr == nil {
			rlog.Error("Session manager not initialized")
			return
		}

		cleanJID := stripDeviceSuffix(v.Info.Chat)
		err := s.whatsappCli.SendChatPresence(cleanJID, types.ChatPresenceComposing, types.ChatPresenceMediaText)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error setting chat presence: %v\n", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), messageTimeout)
		defer cancel()

		if v.Message.GetAudioMessage() != nil {
			audioMsg := v.Message.GetAudioMessage()

			audioData, err := s.whatsappCli.DownloadAny(&waE2E.Message{
				AudioMessage: audioMsg,
			})
			if err != nil {
				rlog.Error("Failed to download audio", "error", err)
				return
			}

			transcription, err := s.openAICli.TranscribeAudio(
				openaicli.TranscribeAudioInput{
					Name: "audio.ogg",
					Data: bytes.NewReader(audioData),
				},
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, "could not transcribe audio: %v\n", err)
				return
			}

			fmt.Println("Transcription:", string(transcription))

			// Process transcription as a regular message
			response, err := s.sessionMgr.SendMessage(
				ctx,
				db,
				s.trelloAPI,
				v.Info.Sender.String(),
				string(transcription),
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error processing transcription: %v\n", err)
				return
			}

			cleanSenderJID := stripDeviceSuffix(v.Info.Sender)
			if _, err := s.whatsappCli.SendMessage(
				context.Background(),
				cleanSenderJID,
				&waE2E.Message{
					Conversation: &response,
				},
			); err != nil {
				fmt.Fprintf(os.Stderr, "could not send message: %v\n", err)
				return
			}
			return
		}

		response, err := s.sessionMgr.SendMessage(
			ctx,
			db,
			s.trelloAPI,
			v.Info.Sender.String(),
			v.Message.GetConversation(),
		)
		if err != nil {
			// Clear typing indicator before returning on error
			if err := s.whatsappCli.SendChatPresence(
				cleanJID,
				types.ChatPresencePaused,
				types.ChatPresenceMediaText,
			); err != nil {
				fmt.Fprintf(os.Stderr, "could not clear chat presence: %v\n", err)
			}
			fmt.Fprintf(os.Stderr, "error processing message: %v\n", err)
			return
		}

		// Clear typing indicator
		if err := s.whatsappCli.SendChatPresence(
			cleanJID,
			types.ChatPresencePaused,
			types.ChatPresenceMediaText,
		); err != nil {
			fmt.Fprintf(os.Stderr, "could not clear chat presence: %v\n", err)
		}

		cleanSenderJID := stripDeviceSuffix(v.Info.Sender)
		if _, err := s.whatsappCli.SendMessage(
			context.Background(),
			cleanSenderJID,
			&waE2E.Message{
				Conversation: &response,
			},
		); err != nil {
			fmt.Fprintf(os.Stderr, "could not send message: %v\n", err)
			return
		}
	}
}

func (s *Service) connectToWhatsApp(deviceStore *store.Device) error {
	s.clientLock.Lock()
	defer s.clientLock.Unlock()

	clientLog := walog.Stdout("whatsapp-client", "INFO", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	client.AddEventHandler(s.whatsappEventHandler)

	s.whatsappCli = client
	s.deviceStore = deviceStore

	if err := client.Connect(); err != nil {
		return fmt.Errorf("could not connect to WhatsApp: %w", err)
	}
	return nil
}

func stripDeviceSuffix(jid types.JID) types.JID {
	return types.JID{
		User:   jid.User,
		Server: jid.Server,
		// Device is intentionally omitted to ensure we're sending to the main user JID
	}
}
