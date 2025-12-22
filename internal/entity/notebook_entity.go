package entity

import (
	"time"

	"github.com/google/uuid"
)

type Notebook struct {
	Id      uuid.UUID
	Name	string
	Parent_id *uuid.UUID
	CreateAt time.Time
	UpdatedAt *time.Time
	DeleteAt *time.Time
	IsDeleted bool
}
