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
		`INSERT INTO note_embedding (id, document, embedding_value, note_id, created_at, updated_at, deleted_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		noteEmbedding.Id,
		noteEmbedding.Document,
		pgvector.NewVector(noteEmbedding.EmbeddingValue),
		noteEmbedding.NoteId,
		noteEmbedding.CreateAt,
		noteEmbedding.UpdatedAt,
		noteEmbedding.DeleteAt,
		noteEmbedding.IsDeleted,
	)

	if err != nil {
		return err
	}

	return nil
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

func NewNoteEmbeddingRepository(db *pgxpool.Pool) INoteEmbeddingRepository {
	return &noteEmbeddingRepository{
		db: db,
	}
}
