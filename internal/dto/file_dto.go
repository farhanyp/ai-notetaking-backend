package dto

import "github.com/google/uuid"

type UploadFileRequest struct {
	NoteId uuid.UUID `form:"note_id" validate:"required"`
}

type UploadFileResponse struct {
	FileId   uuid.UUID `json:"file_id"`
	FileName string    `json:"file_name"`
}
