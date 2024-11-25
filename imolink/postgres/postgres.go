package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

const (
	queryInsertEmbedding string = `INSERT INTO embeddings 
(id, model, text, tokens, vector, created_at) 
VALUES ($1, $2, $3, $4, $5, $6)`

	queryFetchNearestNeighbor string = `SELECT text, vector <-> $1 AS distance
FROM embeddings
ORDER BY distance ASC
LIMIT 1`
)

type (
	StoreEmbeddingInput struct {
		ID        string
		Model     string
		Text      string
		Tokens    int64
		Vector    []float32
		CreatedAt time.Time
	}

	FetchNearestNeighborInput struct{ Vector []float32 }
)

type Postgres struct{ *sqlx.DB }

func NewPostgres(dbConn *sqlx.DB) *Postgres { return &Postgres{DB: dbConn} }

func (p *Postgres) StoreEmbeddings(ctx context.Context, in StoreEmbeddingInput) error {
	if _, err := p.ExecContext(
		ctx, queryInsertEmbedding, in.ID,
		in.Model, in.Text, in.Tokens,
		pgvector.NewVector(in.Vector), in.CreatedAt,
	); err != nil {
		return fmt.Errorf("could not store vector: %w", err)
	}
	return nil
}

func (p *Postgres) FetchNearestNeighbor(ctx context.Context, in FetchNearestNeighborInput) (string, float64, error) {
	var (
		text     string
		distance float64
	)

	row := p.QueryRowxContext(ctx,
		queryFetchNearestNeighbor,
		pgvector.NewVector(in.Vector),
	)

	if err := row.Scan(&text, &distance); err != nil {
		return "", 0, fmt.Errorf("could not fetch nearest neighbor: %w", err)
	}
	return text, distance, nil
}

func (p *Postgres) Purge(ctx context.Context) error {
	if _, err := p.ExecContext(ctx, "DELETE FROM embeddings"); err != nil {
		return fmt.Errorf("could not purge: %w", err)
	}
	return nil
}
