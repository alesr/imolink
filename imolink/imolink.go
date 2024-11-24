package imolink

import (
	"context"
	"fmt"
	"time"

	"encore.app/imolink/postgres"
	"encore.app/internal/pkg/apierror"
	"encore.app/internal/pkg/httpclient"
	"encore.app/internal/pkg/openaicli"
	"encore.dev/beta/errs"
	"encore.dev/pubsub"
	"encore.dev/storage/sqldb"
	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
)

const (
	embbedingModel  = "text-embedding-3-small"
	completionModel = "gpt-4o-mini"
)

var (
	db = sqldb.Named("imolink")

	secrets struct {
		OpenAIKey string
	}

	systemCompletionMsg = openaicli.Message{
		Role: "system",
		Content: `Você é um corretor de imóveis profissional e experiente no mercado imobiliário brasileiro. Seu objetivo é ajudar os clientes a encontrarem o imóvel ideal para suas necessidades.

Diretrizes de comportamento:
- Comunique-se sempre em português brasileiro formal, mas mantenha um tom acolhedor e profissional
- Faça perguntas pertinentes para entender melhor as necessidades do cliente, mas apenas pergunte se for necessário para fornecer uma resposta melhor.
- Evite suposições sobre as preferências do cliente sem ter informações suficientes.
- Seja direto e objetivo nas resposta sempre que houver informações claras disponíveis, evite prolongar a conversa com adicionais perguntas.

Regras de resposta:
1. Quando não houver informações suficientes sobre um imóvel ou característica solicitada, admita que não possui essa informação específica
2. Ao apresentar opções de imóveis:
   - Se houver múltiplos imóveis compatíveis, apresente apenas o mais adequado às necessidades do cliente
   - Forneça uma justificativa sucinta do por que esse imóvel foi selecionado
3. Formato de apresentação do imóvel:
   - Breve descrição do imóvel.

Exemplo de estrutura de resposta com imóvel:
\"[Saudação e contextualização]

*Propriedade:*
Sucinta descrição do imóvel em prosa e comentário sobre adequação do imóvel.

*ID da Propriedade:* REF123\"`,
	}
)

type (
	repository interface {
		FetchNearestNeighbor(ctx context.Context, in postgres.FetchNearestNeighborInput) (string, float64, error)
		StoreEmbeddings(ctx context.Context, in postgres.StoreEmbeddingInput) error
		Purge(ctx context.Context) error
	}

	openAIClient interface {
		CreateEmbedding(ctx context.Context, in openaicli.EmbbedingRequest) (*openaicli.EmbeddingResponse, error)
		CreateChatCompletition(ctx context.Context, in openaicli.CompletitionRequest) (*openaicli.CompletitionResponse, error)
	}

	NewPropertyEvent struct{ Data string }
)

var (
	NewPropertiesTopic = pubsub.NewTopic[*NewPropertyEvent]("new-property", pubsub.TopicConfig{
		DeliveryGuarantee: pubsub.AtLeastOnce,
	})

	_ = pubsub.NewSubscription(
		NewPropertiesTopic,
		"train",
		pubsub.SubscriptionConfig[*NewPropertyEvent]{
			Handler: train,
		},
	)
)

//encore:service
type Service struct {
	repo   repository
	client openAIClient
}

func initService() (*Service, error) {
	return &Service{
		repo: postgres.NewPostgres(
			sqlx.NewDb(db.Stdlib(), "postgres"),
		),
		client: openaicli.New(secrets.OpenAIKey, httpclient.New()),
	}, nil
}

type (
	AskInput  struct{ Question string }
	AskOutput struct{ Answer string }
)

//encore:api private method=POST path=/ask
func (u *Service) Ask(ctx context.Context, in AskInput) (*AskOutput, error) {
	embedd, err := u.client.CreateEmbedding(ctx, openaicli.EmbbedingRequest{
		Model: embbedingModel,
		Input: in.Question,
	})
	if err != nil {
		return nil, apierror.E("could not create question embedding", err, errs.Internal)
	}

	text, _, err := u.repo.FetchNearestNeighbor(ctx, postgres.FetchNearestNeighborInput{
		Vector: embedd.Data[0].Embedding,
	})
	if err != nil {
		return nil, apierror.E("could not fetch nearest neighbor", err, errs.Internal)
	}

	completition, err := u.client.CreateChatCompletition(ctx, openaicli.CompletitionRequest{
		Model: completionModel,
		Messages: []openaicli.Message{
			systemCompletionMsg,
			{
				Role:    "system",
				Content: text,
			},
			{
				Role:    "user",
				Content: in.Question,
			},
		},
	})
	if err != nil {
		return nil, apierror.E("could not create chat completition", err, errs.Internal)
	}
	return &AskOutput{Answer: completition.Choices[0].Message.Content}, nil
}

type TrainingData struct{ Data string }

//encore:api private method=POST path=/train
func (u *Service) Train(ctx context.Context, in TrainingData) error {
	embedd, err := u.client.CreateEmbedding(ctx, openaicli.EmbbedingRequest{
		Model: embbedingModel,
		Input: in.Data,
	})
	if err != nil {
		return apierror.E("could not create training embedding", err, errs.Internal)
	}

	if err := u.repo.StoreEmbeddings(ctx, postgres.StoreEmbeddingInput{
		ID:        ulid.MustNew(ulid.Now(), nil).String(),
		Model:     embbedingModel,
		Text:      in.Data,
		Tokens:    int64(embedd.Usage.TotalTokens),
		Vector:    embedd.Data[0].Embedding,
		CreatedAt: time.Now().UTC(),
	}); err != nil {
		return apierror.E("could not store embeddings", err, errs.Internal)
	}
	return nil
}

//encore:api private method=DELETE path=/embeddings
func (s *Service) Purge(ctx context.Context) error {
	if err := s.repo.Purge(ctx); err != nil {
		return apierror.E("could not purge", err, errs.Internal)
	}
	return nil
}

func train(ctx context.Context, q *NewPropertyEvent) error {
	if err := Train(ctx, TrainingData{Data: q.Data}); err != nil {
		return fmt.Errorf("could not ask: %w", err)
	}
	return nil
}
