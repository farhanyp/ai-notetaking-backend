package entity

import (
	"time"

	"github.com/google/uuid"
)

type NoteEmbedding struct {
	Id             uuid.UUID
	NoteId         uuid.UUID
	FileId         *uuid.UUID
	ChunkContent   string
	EmbeddingValue []float32
	PageNumber     int
	ChunkIndex     int
	OverlapRange   string
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	DeletedAt      *time.Time
	IsDeleted      bool
}
