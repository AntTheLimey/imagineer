-- Imagineer PostgreSQL Extensions
-- This script runs automatically when the database is first initialized.

-- Trigram support for fuzzy text matching
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- pgVector for vector similarity search
CREATE EXTENSION IF NOT EXISTS vector;

-- BM25 full-text search
CREATE EXTENSION IF NOT EXISTS vchord_bm25;

-- Message queue for async processing
CREATE EXTENSION IF NOT EXISTS pgmq;

-- Cron scheduler (required by vectorize)
CREATE EXTENSION IF NOT EXISTS pg_cron;

-- Tokenizer for text processing
CREATE EXTENSION IF NOT EXISTS pg_tokenizer;

-- pgEdge Vectorizer for AI embeddings
CREATE EXTENSION IF NOT EXISTS pgedge_vectorizer;

-- Vectorize for automated embedding pipelines
CREATE EXTENSION IF NOT EXISTS vectorize;
