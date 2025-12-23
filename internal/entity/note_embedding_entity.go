package entity

import (
	"time"

	"github.com/google/uuid"
)

type NoteEmbedding struct {
	Id      uuid.UUID
	Document	string
	EmbeddingValue	[]float32
	NoteId uuid.UUID
	CreateAt time.Time
	UpdatedAt *time.Time
	DeleteAt *time.Time
	IsDeleted bool
}
