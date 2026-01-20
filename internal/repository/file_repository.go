// internal/repository/file_repository.go
package repository

import (
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/pkg/database"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IFileRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) IFileRepository
	Create(ctx context.Context, file *entity.File) error
	GetByNoteId(ctx context.Context, noteId uuid.UUID) (*entity.File, error)
	DeleteByNoteId(ctx context.Context, noteId uuid.UUID) error
	GetByNoteIds(ctx context.Context, noteIds []uuid.UUID) ([]*entity.File, error)
	DeleteByNotebookId(ctx context.Context, notebookId uuid.UUID) error
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

func (r *fileRepository) Create(ctx context.Context, file *entity.File) error {
	_, err := r.db.Exec(
		ctx,
		`INSERT INTO file (id, file_name, original_name, bucket, content_type, note_id, created_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		file.Id,
		file.FileName,
		file.OriginalName,
		file.Bucket,
		file.ContentType,
		file.NoteId,
		file.CreatedAt,
	)
	return err
}

func (r *fileRepository) GetByNoteId(ctx context.Context, noteId uuid.UUID) (*entity.File, error) {
	row := r.db.QueryRow(
		ctx,
		`SELECT id, file_name, original_name, bucket, content_type, note_id, created_at 
         FROM file 
         WHERE note_id = $1`,
		noteId,
	)

	var f entity.File
	// PERBAIKAN: Masukkan f.OriginalName sesuai urutan SELECT (ke-3)
	err := row.Scan(
		&f.Id,
		&f.FileName,
		&f.OriginalName, // Kolom ke-3
		&f.Bucket,
		&f.ContentType,
		&f.NoteId,
		&f.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serverutils.ErrNotFound
		}
		return nil, err
	}
	return &f, nil
}

func (r *fileRepository) DeleteByNoteId(ctx context.Context, noteId uuid.UUID) error {
	_, err := r.db.Exec(
		ctx,
		`DELETE FROM file WHERE note_id = $1`,
		noteId,
	)
	return err
}

func (r *fileRepository) GetByNoteIds(ctx context.Context, noteIds []uuid.UUID) ([]*entity.File, error) {
	// 1. Pastikan semua kolom yang dibutuhkan di-SELECT
	query := `
        SELECT id, file_name, original_name, bucket, note_id 
        FROM file 
        WHERE note_id = ANY($1)
    `

	rows, err := r.db.Query(ctx, query, noteIds)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*entity.File
	for rows.Next() {
		var f entity.File
		// 2. WAJIB: Urutan Scan harus sesuai dengan urutan SELECT di atas
		err := rows.Scan(
			&f.Id,           // id
			&f.FileName,     // file_name (Ini yang jadi 'Key' di S3)
			&f.OriginalName, // original_name
			&f.Bucket,       // bucket
			&f.NoteId,       // note_id
		)

		if err != nil {
			fmt.Printf("[ERROR] Scan error in GetByNoteIds: %v\n", err)
			return nil, err
		}

		files = append(files, &f)
	}

	return files, nil
}

func (r *fileRepository) DeleteByNotebookId(ctx context.Context, notebookId uuid.UUID) error {
	query := `
        DELETE FROM file 
        WHERE note_id IN (
            SELECT id FROM note WHERE notebook_id = $1
        )
    `
	_, err := r.db.Exec(ctx, query, notebookId)
	return err
}
