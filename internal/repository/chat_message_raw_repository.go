package repository

import (
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/pkg/database"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IChatMessageRawRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatMessageRawRepository
	Create(ctx context.Context, chatMessageRaw *entity.ChatMessageRaw) error
	GetChatBySessionId(ctx context.Context, sessionId uuid.UUID) ([]*entity.ChatMessageRaw, error)
	DeleteBySessionId(ctx context.Context, chatSessionId uuid.UUID) error
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
		chatMessageRaw.CreatedAt,
		chatMessageRaw.UpdatedAt,
		chatMessageRaw.DeletedAt,
		chatMessageRaw.IsDeleted,
	)
	if err != nil {
		return err
	}

	return nil
}

func (n *chatmessagerawRepository) GetChatBySessionId(ctx context.Context, sessionId uuid.UUID) ([]*entity.ChatMessageRaw, error) {
	rows, err := n.db.Query(
		ctx,
		`SELECT id, role, chat, session_chat_id, created_at, updated_at, deleted_at, is_deleted FROM chat_message_raw WHERE session_chat_id = $1 AND is_deleted = false ORDER BY created_at ASC`,
		sessionId,
	)

	res := make([]*entity.ChatMessageRaw, 0)

	for rows.Next() {
		var chatMessage entity.ChatMessageRaw
		err := rows.Scan(
			&chatMessage.Id,
			&chatMessage.Role,
			&chatMessage.Chat,
			&chatMessage.ChatSessionId,
			&chatMessage.CreatedAt,
			&chatMessage.UpdatedAt,
			&chatMessage.DeletedAt,
			&chatMessage.IsDeleted,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, &chatMessage)

	}

	return res, err
}

func (n *chatmessagerawRepository) DeleteBySessionId(ctx context.Context, chatSessionId uuid.UUID) error {
	_, err := n.db.Exec(
		ctx,
		`UPDATE chat_message_raw SET deleted_at = $1, is_deleted = true WHERE session_chat_id = $2`,
		time.Now(),
		chatSessionId,
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
