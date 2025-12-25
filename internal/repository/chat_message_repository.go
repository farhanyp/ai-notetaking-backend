package repository

import (
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/pkg/database"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IChatMessageRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatMessageRepository
	Create(ctx context.Context, chatMessage *entity.ChatMessage) error 
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

func NewChatMessageRepository(db *pgxpool.Pool) IChatMessageRepository {
	return &chatmessageRepository{
		db: db,
	}
}
