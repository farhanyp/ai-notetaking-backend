package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateSessionResponse struct {
	Id uuid.UUID `json:"id"`
}

type GetAllSessionResponse struct {
	Id uuid.UUID `json:"id"`
	Title string `json:"title"`
	CreateAt time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type GetChatHistoryResponse struct {
	Id uuid.UUID `json:"id"`
	Role string `json:"role"`
	Chat string `json:"chhat"`
	CreateAt time.Time `json:"created_at"`
}
