package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateNoteRequest struct {
	Title      string    `json:"title" validate:"required"`
	Content    string    `json:"content"`
	NotebookId uuid.UUID `json:"notebook_id" validate:"required"`
}

type CreateNoteResponse struct {
	Id uuid.UUID `json:"id"`
}

type ShowNoteResponse struct {
	Id         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	NotebookId uuid.UUID `json:"notebook_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type UpdateNoteRequest struct {
	Id        uuid.UUID
	Title     string `json:"title" validate:"required"`
	Content   string `json:"content"`
	UpdatedAt time.Time
}

type UpdateNoteResponse struct {
	Id uuid.UUID
}

type MoveNoteRequest struct {
	Id         uuid.UUID
	NotebookId *uuid.UUID `json:"notebook_id"`
}

type MoveNoteResponse struct {
	Id uuid.UUID
}

type SemanticSearchResponse struct {
	Id         uuid.UUID  `json:"id"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	NotebookId uuid.UUID  `json:"notebook_id"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdateAt   *time.Time `json:"updated_at"`
}

type ExtractPreviewResponse struct {
	NoteId        uuid.UUID `json:"note_id"`
	ExtractedText string    `json:"extracted_text"`
}
