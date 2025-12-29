package repository

import (
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/pkg/database"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IChatSessionRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatSessionRepository
	Create(ctx context.Context, chatSession *entity.ChatSession) error
	Update(ctx context.Context, chatSession *entity.ChatSession) error
	Delete(ctx context.Context, sessionId uuid.UUID) error
	GetAllSession(ctx context.Context) ([]*entity.ChatSession, error)
	GetSessionById(ctx context.Context, sessionId uuid.UUID) (*entity.ChatSession, error)
	GetChatBySessionId(ctx context.Context, sessionId uuid.UUID) ([]*entity.ChatMessage, error)
}

type chatbotRepository struct {
	db database.DatabaseQueryer
}

func (n *chatbotRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatSessionRepository {
	return &chatbotRepository{
		db: tx,
	}
}

func (n *chatbotRepository) Create(ctx context.Context, chatSession *entity.ChatSession) error {
	_, err := n.db.Exec(
		ctx,
		`INSERT INTO chat_session (id, title, created_at, updated_at, deleted_at, is_deleted) VALUES ($1, $2, $3, $4, $5, $6)`,
		chatSession.Id,
		chatSession.Title,
		chatSession.CreateAt,
		chatSession.UpdatedAt,
		chatSession.DeleteAt,
		chatSession.IsDeleted,
	)
	if err != nil {
		return err
	}

	return nil
}

func (n *chatbotRepository) GetAllSession(ctx context.Context) ([]*entity.ChatSession, error)  {
	rows, err := n.db.Query(
		ctx,
		`SELECT id, title, created_at, updated_at, deleted_at, is_deleted FROM chat_session WHERE is_deleted = false ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}

	res := make([]*entity.ChatSession, 0)

	for rows.Next() {
		var chatSession entity.ChatSession
		err = rows.Scan(
			&chatSession.Id,
			&chatSession.Title,
			&chatSession.CreateAt,
			&chatSession.UpdatedAt,
			&chatSession.DeleteAt,
			&chatSession.IsDeleted,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, &chatSession)

	}

	return res, err
}

func (n *chatbotRepository) GetSessionById(ctx context.Context, sessionId uuid.UUID) (*entity.ChatSession, error)  {
	rows := n.db.QueryRow(
		ctx,
		`SELECT id, title, created_at, updated_at, deleted_at, is_deleted FROM chat_session WHERE id = $1 AND is_deleted = false`,
		sessionId,
	)

	var chatSession entity.ChatSession
	err := rows.Scan(
		&chatSession.Id,
		&chatSession.Title,
		&chatSession.CreateAt,
		&chatSession.UpdatedAt,
		&chatSession.DeleteAt,
		&chatSession.IsDeleted,
	)

	if err != nil {
		return nil, err
	}


	return &chatSession, err
}

func (n *chatbotRepository) GetChatBySessionId(ctx context.Context, sessionId uuid.UUID) ([]*entity.ChatMessage, error)  {
	rows, err := n.db.Query(
		ctx,
		`SELECT id, role, chat, session_chat_id, created_at, updated_at, deleted_at, is_deleted FROM chat_message WHERE session_chat_id = $1 AND is_deleted = false ORDER BY created_at ASC`,
		sessionId,
	)

	res := make([]*entity.ChatMessage, 0)
	
	for rows.Next() {
		var chatMessage entity.ChatMessage
		err := rows.Scan(
			&chatMessage.Id,
			&chatMessage.Role,
			&chatMessage.Chat,
			&chatMessage.ChatSessionId,
			&chatMessage.CreateAt,
			&chatMessage.UpdatedAt,
			&chatMessage.DeleteAt,
			&chatMessage.IsDeleted,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, &chatMessage)

	}

	return res, err
}

func (n *chatbotRepository) Update(ctx context.Context, chatSession *entity.ChatSession) error {
	_, err := n.db.Exec(
		ctx,
		`UPDATE chat_session SET title = $1, updated_at = $2 WHERE id = $3`,
		chatSession.Title,
		chatSession.UpdatedAt,
		chatSession.Id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (n *chatbotRepository) Delete(ctx context.Context, sessionId uuid.UUID) error {
	_, err := n.db.Exec(
		ctx,
		`UPDATE chat_session SET deleted_at = $1, is_deleted = true WHERE id = $2`,
		time.Now(),
		sessionId,
	)
	if err != nil {
		return err
	}

	return nil
}

func NewChatSessionRepository(db *pgxpool.Pool) IChatSessionRepository {
	return &chatbotRepository{
		db: db,
	}
}
