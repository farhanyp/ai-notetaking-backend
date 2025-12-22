package entity

import (
	"time"

	"github.com/google/uuid"
)

type Note struct {
	Id      uuid.UUID
	Title	string
	Content	string
	Notebook_id uuid.UUID
	CreateAt time.Time
	UpdatedAt *time.Time
	DeleteAt *time.Time
	IsDeleted bool
}
