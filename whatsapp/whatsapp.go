package whatsapp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"encore.app/imolink"
	"encore.app/internal/pkg/openaicli"
	"encore.dev/metrics"
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
	maxInitRetries = 3
	initRetryDelay = 2 * time.Second
	messageTimeout = 2 * time.Minute
)

var (
	db = sqldb.NewDatabase("whatsapp", sqldb.DatabaseConfig{
		Migrations: "./migrations",
	})

	MessagesReceived = metrics.NewCounter[uint64]("messages_received", metrics.CounterConfig{})
	MessagesSent     = metrics.NewCounter[uint64]("messages_sent", metrics.CounterConfig{})

	secrets struct {
		OpenAIKey string
	}
)

// Service is the main service for the WhatsApp API.
//
//encore:service
type Service struct {
	whatsappCli *whatsmeow.Client
	deviceStore *store.Device
	clientLock  sync.Mutex
	sessionMgr  *openaicli.SessionManager
}

func initService() (*Service, error) {
	s := new(Service)

	if err := imolink.InitializeAssistant(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize assistant: %w", err)
	}

	openaiCli := openaicli.New(secrets.OpenAIKey, &http.Client{
		Timeout: 30 * time.Second,
	})

	s.sessionMgr = openaicli.NewSessionManager(
		imolink.Assistant,
		openaiCli,
	)

	dbLog := walog.Stdout("whatsapp-database", "INFO", true)
	container := sqlstore.NewWithDB(db.Stdlib(), "postgres", dbLog)

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		if strings.Contains(err.Error(), "no devices found") {
			return s, nil
		}
		return nil, fmt.Errorf("failed to get WhatsApp device: %w", err)
	}

	if deviceStore != nil {
		if err := s.connectToWhatsApp(deviceStore); err != nil {
			return nil, fmt.Errorf("failed to reconnect to WhatsApp: %w", err)
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

//encore:api auth raw path=/whatsapp/reconnect
func (s *Service) WhatsappReconnect(w http.ResponseWriter, req *http.Request) {
	s.clientLock.Lock()
	defer s.clientLock.Unlock()

	if s.whatsappCli != nil {
		s.whatsappCli.Disconnect()
	}

	dbLog := walog.Stdout("Database", "DEBUG", true)
	container := sqlstore.NewWithDB(db.Stdlib(), "postgres", dbLog)

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		http.Error(w, fmt.Sprintf("No device found: %v", err), http.StatusInternalServerError)
		return
	}

	clientLog := walog.Stdout("Client", "DEBUG", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(s.whatsappEventHandler)

	s.whatsappCli = client
	s.deviceStore = deviceStore

	if err := client.Connect(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to reconnect: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Successfully reconnected to WhatsApp")
}

func (s *Service) whatsappEventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		rlog.Debug(
			"Message received",
			"message", v.Message.GetConversation(),
			"sender", v.Info.Sender,
			"target", v.Info.Sender.User,
		)
		MessagesReceived.Increment()

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

		response, err := s.sessionMgr.SendMessage(ctx, v.Info.Sender.String(), v.Message.GetConversation())
		if err != nil {
			// Clear typing indicator before returning on error
			_ = s.whatsappCli.SendChatPresence(cleanJID, types.ChatPresencePaused, types.ChatPresenceMediaText)
			fmt.Fprintf(os.Stderr, "error processing message: %v\n", err)
			return
		}

		// Clear typing indicator
		if err := s.whatsappCli.SendChatPresence(
			cleanJID,
			types.ChatPresencePaused,
			types.ChatPresenceMediaText,
		); err != nil {
			fmt.Fprintf(os.Stderr, "error clearing chat presence: %v\n", err)
		}

		cleanSenderJID := stripDeviceSuffix(v.Info.Sender)
		if _, err := s.whatsappCli.SendMessage(
			context.Background(),
			cleanSenderJID,
			&waE2E.Message{
				Conversation: &response,
			},
		); err != nil {
			fmt.Fprintf(os.Stderr, "error sending message: %v\n", err)
			return
		}
		MessagesSent.Increment()
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
		return fmt.Errorf("failed to connect WhatsApp client: %w", err)
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
