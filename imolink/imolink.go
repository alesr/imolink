package imolink

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"encore.app/imolink/formatter"
	"encore.app/internal/pkg/apierror"
	"encore.app/properties"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/wiselead-ai/httpclient"
	"github.com/wiselead-ai/openai"
	"github.com/wiselead-ai/trello"
	"github.com/wiselead-ai/whatsapp"
)

const defaultTimeout = time.Minute * 3

var (
	secrets struct {
		OpenAIKey    string
		TrelloAPIKey string
		TrelloToken  string
	}

	db = sqldb.NewDatabase("imolink", sqldb.DatabaseConfig{
		Migrations: "./migrations",
	})

	// Assistant *openaicli.Assistant

	//go:embed assets/*
	assetsFS embed.FS

	//go:embed templates/*
	templatesFS embed.FS
)

type (
	openAIClient interface {
		UploadFile(ctx context.Context, data io.Reader, purpose string) (*openai.FileUploadResponse, error)
		CreateVectorStore(ctx context.Context, in *openai.CreateVectorStoreInput) (*openai.VectorStore, error)
		WaitForVectorStoreCompletion(ctx context.Context, vectorStoreID string, timeout, maxDelay time.Duration) error
		CreateAssistant(ctx context.Context, cfg *openai.CreateAssistantInput) (*openai.Assistant, error)
		ModifyAssistant(ctx context.Context, assistantID string, cfg *openai.ModifyAssistantInput) (*openai.Assistant, error)
		GetAssistant(ctx context.Context, assistantID string) (*openai.Assistant, error)
	}
)

//encore:service
type Service struct {
	client      openAIClient
	mu          sync.RWMutex // to protect assistant updates
	whatsappSvc *whatsapp.Service
	assistantID string
	tmpls       *template.Template
}

func initService() (*Service, error) {
	logger := slog.Default().WithGroup("imolink")

	funcMap := template.FuncMap{
		"safeJS": func(i interface{}) template.JS {
			b, _ := json.Marshal(i)
			return template.JS(b)
		},
	}

	tmpls := template.Must(template.New("dashboard.html").Funcs(funcMap).ParseFS(templatesFS, "templates/dashboard.html"))

	httpCli := httpclient.New(
		httpclient.WithTimeout(defaultTimeout),
	)

	openaiCli := openai.New(logger, secrets.OpenAIKey, httpCli)

	s := &Service{
		client: openaiCli,
		tmpls:  tmpls,
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	// TODO(alesr): gracefully shutdown at some point
	_ = cancel

	if err := s.InitializeAssistant(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize assistant: %w", err)
	}

	trelloCli := trello.NewTrelloAPI(httpCli, secrets.TrelloAPIKey, secrets.TrelloToken)

	wsapp, err := whatsapp.New(logger, db, openaiCli, trelloCli, s.assistantID)
	if err != nil {
		return nil, fmt.Errorf("could not create whatsapp service: %w", err)
	}
	s.whatsappSvc = wsapp
	return s, nil
}

func (s *Service) InitializeAssistant(ctx context.Context) error {
	assistant, err := s.createAssistantWithProperties(ctx)
	if err != nil {
		return apierror.E("failed to initialize assistant", err, errs.Internal)
	}
	s.mu.Lock()
	s.assistantID = assistant.ID
	s.mu.Unlock()
	return nil
}

//encore:api public method=POST path=/imolink/update-assistant
func (s *Service) UpdateAssistant(ctx context.Context) error {
	fileID, err := s.uploadPropertiesFile(ctx)
	if err != nil {
		return apierror.E("failed to upload properties file", err, errs.Internal)
	}

	vectorID, err := s.createVectorStore(ctx, fileID)
	if err != nil {
		return apierror.E("failed to create vector store", err, errs.Internal)
	}

	s.mu.RLock()
	assistantID := s.assistantID
	s.mu.RUnlock()

	assistant, err := s.client.GetAssistant(ctx, assistantID)
	if err != nil {
		return apierror.E("failed to get current assistant", err, errs.Internal)
	}

	_, err = s.client.ModifyAssistant(ctx, assistantID, modifyAssistantCfg(fileID, vectorID, assistant.Instructions))
	if err != nil {
		return apierror.E("failed to modify assistant", err, errs.Internal)
	}

	rlog.Info("Assistant knowledge base updated successfully", "assistantID", assistantID)
	return nil
}

func (s *Service) createAssistantWithProperties(ctx context.Context) (*openai.Assistant, error) {
	fileID, err := s.uploadPropertiesFile(ctx)
	if err != nil {
		return nil, err
	}

	vectorID, err := s.createVectorStore(ctx, fileID)
	if err != nil {
		return nil, err
	}

	assist, err := s.client.CreateAssistant(ctx, assistantCfg(fileID, vectorID))
	if err != nil {
		return nil, fmt.Errorf("could not create assistant: %w", err)
	}
	rlog.Info("Assistant created", "assistantID", assist.ID)
	return assist, nil
}

func (s *Service) uploadPropertiesFile(ctx context.Context) (string, error) {
	props, err := properties.List(ctx, properties.ListInput{})
	if err != nil {
		return "", fmt.Errorf("could not list properties: %w", err)
	}

	if len(props.Properties) == 0 {
		return "", fmt.Errorf("no properties available in the database")
	}

	rlog.Info("Uploading properties data to OpenAI", "properties", len(props.Properties))

	file, err := s.client.UploadFile(
		ctx,
		strings.NewReader(
			formatter.FormatProperties(props.Properties),
		),
		"assistants",
	)
	if err != nil {
		return "", fmt.Errorf("could not upload properties data: %w", err)
	}

	rlog.Info("Properties data uploaded to OpenAI", "fileID", file.ID)
	return file.ID, nil
}

func (s *Service) createVectorStore(ctx context.Context, fileID string) (string, error) {
	vector, err := s.client.CreateVectorStore(ctx,
		&openai.CreateVectorStoreInput{
			Name:    "properties",
			FileIDs: []string{fileID},
		},
	)
	if err != nil {
		return "", fmt.Errorf("could not create vector store: %w", err)
	}

	if vector.Status != "completed" {
		if err := s.client.WaitForVectorStoreCompletion(
			ctx,
			vector.ID,
			defaultTimeout,
			15*time.Second,
		); err != nil {
			return "", fmt.Errorf("could not wait for vector store completion: %w", err)
		}
	}
	rlog.Info("Vector store created", "vectorStoreID", vector.ID)
	return vector.ID, nil
}

func assistantCfg(fileID, vectorStoreID string) *openai.CreateAssistantInput {
	return &openai.CreateAssistantInput{
		Name:         "ImoLink",
		Description:  "Assistente especializado em imóveis em Aracaju",
		Model:        openai.AssistantModel,
		Instructions: assistantInstructions,
		Tools: []openai.Tool{
			{Type: openai.ToolTypeFileSearch},
			{Type: openai.ToolTypeCodeInterpreter},
			{
				Type:     openai.ToolTypeFunction,
				Function: leadFunctionDefinition(),
			},
		},
		ToolResources: openai.ToolResources{
			CodeInterpreter: &openai.CodeInterpreter{FileIDs: []string{fileID}},
			FileSearch:      &openai.FileSearch{VectorStoreIDs: []string{vectorStoreID}},
		},
		Metadata: openai.Meta{
			"type":    "real_estate_assistant",
			"region":  "Aracaju",
			"version": "1.0",
		},
	}
}

func modifyAssistantCfg(fileID, vectorStoreID, currentInstructions string) *openai.ModifyAssistantInput {
	return &openai.ModifyAssistantInput{
		Description:  "Assistente especializado em imóveis em Aracaju - Atualizado",
		Instructions: currentInstructions,
		Tools: []openai.Tool{
			{Type: openai.ToolTypeFileSearch},
			{Type: openai.ToolTypeCodeInterpreter},
			{
				Type:     openai.ToolTypeFunction,
				Function: leadFunctionDefinition(),
			},
		},
		ToolResources: openai.ToolResources{
			CodeInterpreter: &openai.CodeInterpreter{FileIDs: []string{fileID}},
			FileSearch:      &openai.FileSearch{VectorStoreIDs: []string{vectorStoreID}},
		},
		Metadata: openai.Meta{
			"type":    "real_estate_assistant",
			"region":  "Aracaju",
			"version": "1.1",
		},
	}
}

func leadFunctionDefinition() *openai.FunctionDefinition {
	return &openai.FunctionDefinition{
		Name:        "lead",
		Description: "Create a new lead in the system. This function MUST be called when the user provides their name.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "The name of the lead/user",
					"minLength":   1,
				},
			},
			"required": []string{"name"},
		},
	}
}

//encore:api public raw path=/whatsapp/connect
func (s *Service) WhatsappConnect(w http.ResponseWriter, req *http.Request) {
	if err := s.whatsappSvc.WhatsappConnect(w); err != nil {
		apierror.E("could not connect to WhatsApp", err, errs.Internal)
	}
}
