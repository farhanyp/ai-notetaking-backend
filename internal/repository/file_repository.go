// internal/repository/file_repository.go
package repository

import (
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/pkg/database"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IFileRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) IFileRepository
	Create(ctx context.Context, file *entity.File) error
	GetByNoteId(ctx context.Context, noteId uuid.UUID) (*entity.File, error)
	DeleteByNoteId(ctx context.Context, noteId uuid.UUID) error
}

type fileRepository struct {
	db database.DatabaseQueryer
}

func NewFileRepository(db *pgxpool.Pool) IFileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) IFileRepository {
	return &fileRepository{db: tx}
}

// Create menyimpan metadata file baru
func (r *fileRepository) Create(ctx context.Context, file *entity.File) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO file (id, file_name, bucket, content_type, note_id, created_at) 
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		file.Id,
		file.FileName,
		file.Bucket,
		file.ContentType,
		file.NoteId,
		file.CreatedAt,
	)
	return err
}

// GetByNoteId mengambil file yang menempel pada satu Note
func (r *fileRepository) GetByNoteId(ctx context.Context, noteId uuid.UUID) (*entity.File, error) {
	row := r.db.QueryRow(
		ctx,
		`SELECT id, file_name, bucket, content_type, note_id, created_at 
		 FROM file 
		 WHERE note_id = $1`,
		noteId,
	)

	var f entity.File
	err := row.Scan(&f.Id, &f.FileName, &f.Bucket, &f.ContentType, &f.NoteId, &f.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serverutils.ErrNotFound
		}
		return nil, err
	}
	return &f, nil
}

// DeleteByNoteId melakukan soft delete pada file berdasarkan note_id
func (r *fileRepository) DeleteByNoteId(ctx context.Context, noteId uuid.UUID) error {
	_, err := r.db.Exec(
		ctx,
		`UPDATE file SET is_deleted = true, deleted_at = $1 WHERE note_id = $2`,
		time.Now(),
		noteId,
	)
	return err
}
