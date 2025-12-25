package dto

import "github.com/google/uuid"

type CreateSessionResponse struct {
	Id uuid.UUID `json:"id"`
}
