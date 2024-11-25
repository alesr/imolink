package whatsapp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	"encore.app/imolink"
	"encore.dev"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	walog "go.mau.fi/whatsmeow/util/log"
)

var (
	db = sqldb.NewDatabase("whatsapp", sqldb.DatabaseConfig{
		Migrations: "./migrations",
	})

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
}

func initService() (*Service, error) {
	dbLog := walog.Stdout("whatsapp-database", "INFO", true)
	container := sqlstore.NewWithDB(db.Stdlib(), "postgres", dbLog)

	s := new(Service)

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

//encore:api auth raw path=/whatsapp/connect
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

	// Create a new client
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

// WhatsappReconnect is an API method to reconnect to WhatsApp.
//
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

		resp, err := imolink.AskQuestion(
			context.Background(),
			imolink.QuestionInput{
				Question: v.Message.GetConversation(),
			},
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error asking model: %v\n", err)
			return
		}

		if _, err := s.whatsappCli.SendMessage(
			context.Background(),
			v.Info.Sender,
			&waE2E.Message{
				Conversation: &resp.Answer,
			},
		); err != nil {
			fmt.Fprintf(os.Stderr, "error sending message: %v\n", err)
			return
		}

		ref := extractPropertyRef(resp.Answer)
		if ref == "" {
			return
		}

		url := fmt.Sprintf(
			"%s://%s/properties/%s",
			encore.Meta().APIBaseURL.Scheme,
			encore.Meta().APIBaseURL.Host,
			ref,
		)

		if _, err := s.whatsappCli.SendMessage(
			context.Background(),
			v.Info.Sender,
			&waE2E.Message{
				Conversation: &url,
			},
		); err != nil {
			fmt.Fprintf(os.Stderr, "error sending property URL: %v\n", err)
			return
		}
	}
}

func extractPropertyRef(message string) string {
	patterns := []string{
		`REF\d+`,
		`ID da Propriedade:\s*(\S+)`,
		`Reference:\s*(\S+)`,
		`Referência:\s*(\S+)`,
	}

	message = strings.ReplaceAll(message, "\n", " ")

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindString(message); match != "" {
			ref := strings.TrimSpace(match)

			if strings.Contains(ref, ":") {
				parts := strings.Split(ref, ":")
				ref = strings.TrimSpace(parts[len(parts)-1])
			}
			return ref
		}
	}
	return ""
}
