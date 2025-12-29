package repository

import (
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/pkg/database"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IChatMessageRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatMessageRepository
	Create(ctx context.Context, chatMessage *entity.ChatMessage) error
	DeleteBySessionId(ctx context.Context, chatSessionId uuid.UUID) error
}

type chatmessageRepository struct {
	db database.DatabaseQueryer
}

func (n *chatmessageRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatMessageRepository {
	return &chatmessageRepository{
		db: tx,
	}
}

func (n *chatmessageRepository) Create(ctx context.Context, chatMessage *entity.ChatMessage) error {
	_, err := n.db.Exec(
		ctx,
		`INSERT INTO chat_message (id, role, chat, session_chat_id, created_at, updated_at, deleted_at, is_deleted) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		chatMessage.Id,
		chatMessage.Role,
		chatMessage.Chat,
		chatMessage.ChatSessionId,
		chatMessage.CreateAt,
		chatMessage.UpdatedAt,
		chatMessage.DeleteAt,
		chatMessage.IsDeleted,
	)
	if err != nil {
		return err
	}

	return nil
}

func (n *chatmessageRepository) DeleteBySessionId(ctx context.Context, chatSessionId uuid.UUID) error {
	_, err := n.db.Exec(
		ctx,
		`UPDATE chat_message SET deleted_at = $1, is_deleted = true WHERE session_chat_id = $2`,
		time.Now(),
		chatSessionId,
	)
	if err != nil {
		return err
	}

	return nil
}

func NewChatMessageRepository(db *pgxpool.Pool) IChatMessageRepository {
	return &chatmessageRepository{
		db: db,
	}
}
