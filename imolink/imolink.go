package imolink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"encore.app/imolink/formatter"
	"encore.app/internal/pkg/apierror"
	"encore.app/internal/pkg/httpclient"
	"encore.app/internal/pkg/openaicli"
	"encore.app/internal/pkg/openaicli/types"
	"encore.app/properties"
	"encore.dev/beta/errs"
	"encore.dev/metrics"
	"encore.dev/pubsub"
)

var (
	// Encore metric collectors
	QuestionsProcessed = metrics.NewCounter[uint64]("questions_processed", metrics.CounterConfig{})
	DataTrained        = metrics.NewCounter[uint64]("data_trained", metrics.CounterConfig{})

	NewPropertiesTopic = pubsub.NewTopic[*NewPropertyEvent]("new-property", pubsub.TopicConfig{
		DeliveryGuarantee: pubsub.AtLeastOnce,
	})

	_ = pubsub.NewSubscription(
		NewPropertiesTopic,
		"new-property",
		pubsub.SubscriptionConfig[*NewPropertyEvent]{
			Handler: train,
		},
	)

	secrets struct {
		OpenAIKey string
	}
)

const (
	defaultTimeout     = 180 * time.Second
	defaultDialTimeout = 30 * time.Second
	defaultTLSTimeout  = 20 * time.Second
	defaultKeepAlive   = 90 * time.Second
)

var (
	Assistant *types.Assistant
)

type (
	openAIClient interface {
		CreateAssistant(ctx context.Context, cfg types.AssistantCfg) (*types.Assistant, error)
		NewThread(ctx context.Context) (*types.Thread, error)
		UploadFile(ctx context.Context, data io.Reader, purpose string) (*types.FileUploadResponse, error)
		AddMessage(ctx context.Context, in types.CreateMessageInput) error
		RunThread(ctx context.Context, threadID, assistantID string) (*types.Run, error)
		GetMessages(ctx context.Context, threadID string) (*types.ThreadMessageList, error)
		AttachFileToAssistant(ctx context.Context, assistantID, fileID string) error
		CreateVectorStore(ctx context.Context, in *openaicli.CreateVectorStoreRequest) (*openaicli.VectorStoreResponse, error)
		WaitForVectorStoreCompletion(ctx context.Context, vectorStoreID string, timeout, maxDelay time.Duration) error
	}

	NewPropertyEvent  struct{ Data string }
	QuestionsInput    struct{ Input []QuestionInput }
	QuestionInput     struct{ Role, Question string }
	QuestionOutput    struct{ Answer string }
	TrainingDataInput struct{ Data []string }
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

func (s *Service) initializeAssistantWithProperties(ctx context.Context) (*types.Assistant, error) {
	props, err := properties.List(ctx, properties.ListInput{})
	if err != nil {
		return nil, fmt.Errorf("could not list properties: %w", err)
	}

	if len(props.Properties) == 0 {
		return nil, fmt.Errorf("no properties available in the database")
	}

	data := formatter.FormatProperties(props.Properties)

	fileResp, err := s.client.UploadFile(ctx, strings.NewReader(data), "assistants")
	if err != nil {
		return nil, fmt.Errorf("could not upload properties data: %w", err)
	}

	in := openaicli.CreateVectorStoreRequest{
		Name:    "properties",
		FileIDs: []string{fileResp.ID},
	}

	resp, err := s.client.CreateVectorStore(ctx, &in)
	if err != nil {
		return nil, fmt.Errorf("could not create vector store file: %w", err)
	}

	if resp.Status != "completed" {
		if err := s.client.WaitForVectorStoreCompletion(ctx, resp.ID, defaultTimeout, 10*time.Second); err != nil {
			return nil, fmt.Errorf("could not wait for vector store completion: %w", err)
		}
	}

	assist, err := s.client.CreateAssistant(ctx, types.AssistantCfg{
		Name:         "ImoLink",
		Description:  "Assistente especializado em imóveis em Aracaju",
		Model:        types.AssistantModel,
		Instructions: assistantInstructions,
		Tools: []types.Tool{
			{Type: types.ToolTypeFileSearch},
			{Type: types.ToolTypeCodeInterpreter},
		},
		ToolResources: types.ToolResources{
			CodeInterpreter: &types.CodeInterpreter{FileIDs: []string{fileResp.ID}},
			FileSearch:      &types.FileSearch{VectorStoreIDs: []string{resp.ID}},
		},
		Metadata: types.Meta{
			"type":    "real_estate_assistant",
			"region":  "Aracaju",
			"version": "1.0",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("could not create assistant: %w", err)
	}
	return assist, nil
}

//encore:api public method=POST path=/imolink/training-data
func (s *Service) AddTrainingData(ctx context.Context, in TrainingDataInput) error {
	s.mu.RLock()
	assistant := Assistant
	s.mu.RUnlock()

	if assistant == nil {
		return apierror.E("assistant not initialized", nil, errs.Internal)
	}

	jsonData, err := json.Marshal(in.Data)
	if err != nil {
		return apierror.E("invalid json string", err, errs.InvalidArgument)
	}

	resp, err := s.client.UploadFile(ctx, bytes.NewReader(jsonData), "assistants")
	if err != nil {
		return apierror.E("could not upload file", err, errs.Internal)
	}

	if err := s.client.AttachFileToAssistant(ctx, assistant.ID, resp.ID); err != nil {
		return apierror.E("could not attach file to assistant", err, errs.Internal)
	}

	DataTrained.Increment()
	return nil
}

func train(ctx context.Context, q *NewPropertyEvent) error {
	if err := AddTrainingData(ctx, TrainingDataInput{Data: []string{q.Data}}); err != nil {
		return fmt.Errorf("could not ask: %w", err)
	}
	return nil
}
