package entity

import (
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	Id uuid.UUID
	Role string
	Chat string
	ChatSessionId uuid.UUID
	CreateAt time.Time
	UpdatedAt *time.Time
	DeleteAt *time.Time
	IsDeleted bool
}
