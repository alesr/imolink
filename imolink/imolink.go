package imolink

import (
	"context"
	"embed"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"encore.app/imolink/formatter"
	"encore.app/internal/pkg/apierror"
	"encore.app/internal/pkg/httpclient"
	"encore.app/internal/pkg/openaicli"
	"encore.app/properties"
	"encore.dev/beta/errs"
)

const defaultTimeout = time.Minute

var (
	Assistant *openaicli.Assistant

	//go:embed assets/*
	assetsFS embed.FS

	secrets struct {
		OpenAIKey string
	}
)

type (
	openAIClient interface {
		UploadFile(ctx context.Context, data io.Reader, purpose string) (*openaicli.FileUploadResponse, error)
		CreateVectorStore(ctx context.Context, in *openaicli.CreateVectorStoreInput) (*openaicli.VectorStore, error)
		WaitForVectorStoreCompletion(ctx context.Context, vectorStoreID string, timeout, maxDelay time.Duration) error
		CreateAssistant(ctx context.Context, cfg *openaicli.CreateAssistantInput) (*openaicli.Assistant, error)
	}
)

//encore:service
type Service struct {
	client openAIClient
	mu     sync.RWMutex // to protect assistant updates
}

func initService() (*Service, error) {
	return &Service{
		client: openaicli.New(
			secrets.OpenAIKey,
			httpclient.New(
				httpclient.WithTimeout(defaultTimeout),
			),
		),
	}, nil
}

//encore:api public method=POST path=/imolink/init-assistant
func (s *Service) InitializeAssistant(ctx context.Context) error {
	assistant, err := s.initializeAssistantWithProperties(ctx)
	if err != nil {
		return apierror.E("failed to initialize assistant", err, errs.Internal)
	}

	s.mu.Lock()
	Assistant = assistant
	s.mu.Unlock()
	return nil
}

func (s *Service) initializeAssistantWithProperties(ctx context.Context) (*openaicli.Assistant, error) {
	// We fetch the properties from the db and  upload the data
	// to openai so that we can use it with the code interpreter tool.

	props, err := properties.List(ctx, properties.ListInput{})
	if err != nil {
		return nil, fmt.Errorf("could not list properties: %w", err)
	}

	if len(props.Properties) == 0 {
		return nil, fmt.Errorf("no properties available in the database")
	}

	uploadedFile, err := s.client.UploadFile(
		ctx,
		strings.NewReader(
			formatter.FormatProperties(props.Properties),
		),
		"assistants",
	)
	if err != nil {
		return nil, fmt.Errorf("could not upload properties data: %w", err)
	}

	// Once we have the file uploaded, we create a vector store.

	vectorStore, err := s.client.CreateVectorStore(ctx,
		&openaicli.CreateVectorStoreInput{
			Name:    "properties",
			FileIDs: []string{uploadedFile.ID},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not create vector store file: %w", err)
	}

	if vectorStore.Status != "completed" {
		if err := s.client.WaitForVectorStoreCompletion(
			ctx,
			vectorStore.ID,
			defaultTimeout,
			10*time.Second,
		); err != nil {
			return nil, fmt.Errorf("could not wait for vector store completion: %w", err)
		}
	}

	assist, err := s.client.CreateAssistant(ctx, assistantCfg(uploadedFile.ID, vectorStore.ID))
	if err != nil {
		return nil, fmt.Errorf("could not create assistant: %w", err)
	}
	return assist, nil
}

func assistantCfg(fileID, vectorStoreID string) *openaicli.CreateAssistantInput {
	return &openaicli.CreateAssistantInput{
		Name:         "ImoLink",
		Description:  "Assistente especializado em imóveis em Aracaju",
		Model:        openaicli.AssistantModel,
		Instructions: assistantInstructions,
		Tools: []openaicli.Tool{
			{Type: openaicli.ToolTypeFileSearch},
			{Type: openaicli.ToolTypeCodeInterpreter},
			{
				Type:     openaicli.ToolTypeFunction,
				Function: leadFunctionDefinition(),
			},
		},
		ToolResources: openaicli.ToolResources{
			CodeInterpreter: &openaicli.CodeInterpreter{FileIDs: []string{fileID}},
			FileSearch:      &openaicli.FileSearch{VectorStoreIDs: []string{vectorStoreID}},
		},
		Metadata: openaicli.Meta{
			"type":    "real_estate_assistant",
			"region":  "Aracaju",
			"version": "1.0",
		},
	}
}

func leadFunctionDefinition() *openaicli.FunctionDefinition {
	return &openaicli.FunctionDefinition{
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
