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
	CreatedAt time.Time `json:"created_at"`
}

type SendChatResponseChat struct{
	Id uuid.UUID `json:"id"`
	Chat string `json:"chat"`
	Role string `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type SendChatRequest struct{
	ChatSessionId uuid.UUID `json:"chat_session_id" validate:"required"`
	Chat string `json:"chat" validate:"required"`
}

type SendChatResponse struct {
	ChatSessionId uuid.UUID `json:"chat_session_id"`
	ChatSessionTitle string `json:"chat_session_title"`
	Send *SendChatResponseChat `json:"send"`
	Reply *SendChatResponseChat `json:"reply"`
}

type DeleteSessionRequest struct {
	ChatSessionId uuid.UUID `json:"chat_session_id"`
}
