package dto

import "github.com/google/uuid"

type PublishEmbedNoteMessage struct {
	NotedId uuid.UUID `json:"note_id"`
}