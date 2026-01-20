package repository

import (
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/pkg/database"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type INoteEmbeddingRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) INoteEmbeddingRepository
	Create(ctx context.Context, noteEmbedding *entity.NoteEmbedding) error
	DeleteByID(ctx context.Context, noteId uuid.UUID) error
	SemanticSearch(ctx context.Context, embeddingValues []float32) ([]*entity.NoteEmbedding, error)
	DeleteByNotebookId(ctx context.Context, notebookId uuid.UUID) error
	SearchSimilarity(ctx context.Context, embeddingValues []float32) ([]*entity.NoteEmbedding, error)
}

type noteEmbeddingRepository struct {
	db database.DatabaseQueryer
}

func (n *noteEmbeddingRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) INoteEmbeddingRepository {
	return &noteEmbeddingRepository{
		db: tx,
	}
}

func (n *noteEmbeddingRepository) Create(ctx context.Context, noteEmbedding *entity.NoteEmbedding) error {
	_, err := n.db.Exec(
		ctx,
		`INSERT INTO note_embedding (id, note_id, file_id, chunk_content, embedding_value, page_number, chunk_index, overlap_range, created_at, updated_at, deleted_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		noteEmbedding.Id,
		noteEmbedding.NoteId,
		noteEmbedding.FileId,
		noteEmbedding.ChunkContent,
		pgvector.NewVector(noteEmbedding.EmbeddingValue),
		noteEmbedding.PageNumber,
		noteEmbedding.ChunkIndex,
		noteEmbedding.OverlapRange,
		noteEmbedding.CreatedAt,
		noteEmbedding.UpdatedAt,
		noteEmbedding.DeletedAt,
		noteEmbedding.IsDeleted,
	)

	if err != nil {
		return err
	}

	return nil
}

func (n *noteEmbeddingRepository) SemanticSearch(ctx context.Context, embeddingValues []float32) ([]*entity.NoteEmbedding, error) {
	rows, err := n.db.Query(
		ctx,
		`SELECT id, note_id from note_embedding WHERE is_deleted = false ORDER BY 1 - (embedding_value <-> $1) DESC LIMIT 5`,
		pgvector.NewVector(embeddingValues),
	)
	if err != nil {
		return nil, err
	}

	res := make([]*entity.NoteEmbedding, 0)
	for rows.Next() {
		var noteEmbedding entity.NoteEmbedding
		err := rows.Scan(
			&noteEmbedding.Id,
			&noteEmbedding.NoteId,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, &noteEmbedding)
	}

	return res, nil
}

func (n *noteEmbeddingRepository) DeleteByID(ctx context.Context, noteId uuid.UUID) error {
	_, err := n.db.Exec(
		ctx,
		`UPDATE note_embedding SET deleted_at = $1, is_deleted = true WHERE note_id = $2`,
		time.Now(),
		noteId,
	)

	if err != nil {
		return err
	}

	return nil
}

func (n *noteEmbeddingRepository) DeleteByNotebookId(ctx context.Context, notebookId uuid.UUID) error {
	_, err := n.db.Exec(
		ctx,
		`UPDATE note_embedding SET deleted_at = $1, is_deleted = true WHERE note_id IN (SELECT id FROM note WHERE notebook_id = $2 AND is_deleted = false)`,
		time.Now(),
		notebookId,
	)

	if err != nil {
		return err
	}

	return nil
}

func (n *noteEmbeddingRepository) SearchSimilarity(ctx context.Context, embeddingValues []float32) ([]*entity.NoteEmbedding, error) {
	rows, err := n.db.Query(
		ctx,
		`SELECT id, document from note_embedding WHERE is_deleted = false ORDER BY 1 - (embedding_value <-> $1) DESC LIMIT 5`,
		pgvector.NewVector(embeddingValues),
	)
	if err != nil {
		return nil, err
	}

	res := make([]*entity.NoteEmbedding, 0)
	for rows.Next() {
		var noteEmbedding entity.NoteEmbedding
		err := rows.Scan(
			&noteEmbedding.Id,
			&noteEmbedding.ChunkContent,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, &noteEmbedding)
	}

	return res, nil
}

func NewNoteEmbeddingRepository(db *pgxpool.Pool) INoteEmbeddingRepository {
	return &noteEmbeddingRepository{
		db: db,
	}
}
