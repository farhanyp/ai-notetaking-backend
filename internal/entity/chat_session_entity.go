package entity

import (
	"time"

	"github.com/google/uuid"
)

type ChatSession struct {
	Id uuid.UUID
	Title string
	CreateAt time.Time
	UpdatedAt *time.Time
	DeleteAt *time.Time
	IsDeleted bool
}
