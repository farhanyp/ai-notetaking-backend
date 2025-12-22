package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateNoteRequest struct {
	Title 		string 		`json:"title" validate:"required"`
	Content 	string 		`json:"content"`
	Notebook_id uuid.UUID 	`json:"notebook_id" validate:"required"`
}

type CreateNoteResponse struct {
	Id uuid.UUID `json:"id"`
}

type ShowNoteResponse struct {
	Id      uuid.UUID `json:"id"`
	Title	string `json:"title"`
	Content string `json:"content"`
	Notebook_id uuid.UUID `json:"notebook_id"`
	CreateAt time.Time `json:"created_at"`
}

type UpdateNoteRequest struct {
	Id uuid.UUID
	Title 		string 		`json:"title" validate:"required"`
	Content 	string 		`json:"content"`
	Updated_at time.Time
}

type UpdateNoteResponse struct {
	Id uuid.UUID
}

type MoveNoteRequest struct {
	Id uuid.UUID
	Notebook_id *uuid.UUID `json:"notebook_id"`
}

type MoveNoteResponse struct {
	Id uuid.UUID
}
