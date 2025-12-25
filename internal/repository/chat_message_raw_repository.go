package repository

import (
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/pkg/database"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IChatMessageRawRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatMessageRawRepository
	Create(ctx context.Context, chatMessageRaw *entity.ChatMessageRaw) error
}

type chatmessagerawRepository struct {
	db database.DatabaseQueryer
}

func (n *chatmessagerawRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatMessageRawRepository {
	return &chatmessagerawRepository{
		db: tx,
	}
}

func (n *chatmessagerawRepository) Create(ctx context.Context, chatMessageRaw *entity.ChatMessageRaw) error {
	_, err := n.db.Exec(
		ctx,
		`INSERT INTO chat_message_raw (id, role, chat, session_chat_id, created_at, updated_at, deleted_at, is_deleted) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		chatMessageRaw.Id,
		chatMessageRaw.Role,
		chatMessageRaw.Chat,
		chatMessageRaw.ChatSessionId,
		chatMessageRaw.CreateAt,
		chatMessageRaw.UpdatedAt,
		chatMessageRaw.DeleteAt,
		chatMessageRaw.IsDeleted,
	)
	if err != nil {
		return err
	}

	return nil
}

func NewChatMessageRawRepository(db *pgxpool.Pool) IChatMessageRawRepository {
	return &chatmessagerawRepository{
		db: db,
	}
}
