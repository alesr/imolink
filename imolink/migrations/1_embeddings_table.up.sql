CREATE EXTENSION vector;

CREATE TABLE embeddings (
    id  VARCHAR(255) PRIMARY KEY, 
    model VARCHAR(255) NOT NULL,
    text TEXT NOT NULL,
    tokens INTEGER NOT NULL,
    vector vector(1536) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX ON embeddings USING hnsw (vector vector_cosine_ops);
