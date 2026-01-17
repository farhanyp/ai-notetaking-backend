package entity

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	Id          uuid.UUID
	FileName    string
	Bucket      string
	ContentType string
	NoteId      uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	DeletedAt   *time.Time
	IsDeleted   bool
}
