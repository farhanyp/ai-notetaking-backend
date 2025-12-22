package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateNotebookRequest struct {
	Name string `json:"name" validate:"required"`
	Parent_id *uuid.UUID `json:"parent_id"`
}

type CreateNotebookResponse struct {
	Id uuid.UUID `json:"id"`
}

type UpdateNotebookRequest struct {
	Id uuid.UUID
	Name string `json:"name" validate:"required"`
}

type UpdateNotebookResponse struct {
	Id uuid.UUID `json:"id"`
}

type MoveNotebookRequest struct {
	Id uuid.UUID
	Parent_id *uuid.UUID `json:"Parent_id"`
}

type MoveNotebookResponse struct {
	Id uuid.UUID `json:"id"`
}

type ShowNotebookResponse struct {
	Id      uuid.UUID `json:"id"`
	Name	string `json:"name"`
	Parent_id *uuid.UUID `json:"parent_id"`
	CreateAt time.Time `json:"created_at"`
}

type ListNotebookResponse struct {
	Id      uuid.UUID `json:"id"`
	Name	string `json:"name"`
	Parent_id *uuid.UUID `json:"parent_id"`
	CreateAt time.Time `json:"created_at"`
	UpdateAt *time.Time `json:"updated_at"`

	Notes []*GetAllNotebookResponseNote `json:"notes"`
}

type GetAllNotebookResponseNote struct {
	Id      uuid.UUID `json:"id"`
	Title	string `json:"title"`
	Content	string `json:"content"`
	CreateAt time.Time `json:"created_at"`
	UpdateAt *time.Time `json:"updated_at"`
}
