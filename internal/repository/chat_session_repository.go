package repository

import (
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/pkg/database"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IChatSessionRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatSessionRepository
	Create(ctx context.Context, chatSession *entity.ChatSession) error 
	GetAllSession(ctx context.Context) ([]*entity.ChatSession, error)
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

func NewChatSessionRepository(db *pgxpool.Pool) IChatSessionRepository {
	return &chatbotRepository{
		db: db,
	}
}
