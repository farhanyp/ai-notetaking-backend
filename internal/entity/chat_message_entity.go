package entity

import (
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	Id            uuid.UUID
	Role          string
	Chat          string
	ChatSessionId uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     *time.Time
	DeletedAt     *time.Time
	IsDeleted     bool
}
