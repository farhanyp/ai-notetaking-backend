package service

import (
	"ai-notetaking-be/internal/constant"
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/repository"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IChatbotService interface {
	CreateSession(ctx context.Context) (*dto.CreateSessionResponse, error)
	GetAllSession(ctx context.Context) ([]*dto.GetAllSessionResponse, error)
}

type chatbotService struct {
	db *pgxpool.Pool
	chatSessionRepository repository.IChatSessionRepository
	chatMessageRepository    repository.IChatMessageRepository
	chatMessageRawRepository repository.IChatMessageRawRepository
}

func NewChatbotService(db *pgxpool.Pool, chatSessionRepository repository.IChatSessionRepository, chatMessageRepository repository.IChatMessageRepository, chatMessageRawRepository repository.IChatMessageRawRepository) IChatbotService {
	return &chatbotService{
		db: db,
		chatSessionRepository:    chatSessionRepository,
		chatMessageRepository:    chatMessageRepository,
		chatMessageRawRepository: chatMessageRawRepository,
	}
}

func (c *chatbotService) CreateSession(ctx context.Context) (*dto.CreateSessionResponse, error){

	now := time.Now()
	chatSession := &entity.ChatSession{
		Id: uuid.New(),
		Title: "Unamed session",
		CreateAt: now,
	}

	chatMessage := &entity.ChatMessage{
		Id: uuid.New(),
		Chat: "Hi, how can i help you ?",
		Role: constant.ChatMessageRoleModel,
		ChatSessionId: chatSession.Id,
		CreateAt: now,
	}

	chatMessageRawUser := &entity.ChatMessageRaw{
		Id: uuid.New(),
		Chat: constant.ChatMessageRawInititalUserPromptV1,
		Role: constant.ChatMessageRoleUser,
		ChatSessionId: chatSession.Id,
		CreateAt: now,
	}

	chatMessageRawModel := &entity.ChatMessageRaw{
		Id: uuid.New(),
		Chat: constant.ChatMessageRawInititalModelPromptV1,
		Role: constant.ChatMessageRoleModel,
		ChatSessionId: chatSession.Id,
		CreateAt: now,
	}

	tx, err := c.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	chatSessionRepository := c.chatSessionRepository.UsingTx(ctx, tx)
	chatMessageRepository := c.chatMessageRepository.UsingTx(ctx, tx)
	chatMessageRawRepository := c.chatMessageRawRepository.UsingTx(ctx, tx)

	err = chatSessionRepository.Create(ctx, chatSession)
	if err != nil {
		return nil, err
	}

	err = chatMessageRepository.Create(ctx, chatMessage)
	if err != nil {
		return nil, err
	}

	err = chatMessageRawRepository.Create(ctx, chatMessageRawUser)
	if err != nil {
		return nil, err
	}

	err = chatMessageRawRepository.Create(ctx, chatMessageRawModel)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.CreateSessionResponse{
		Id: chatSession.Id,
	}, nil

}

func (c *chatbotService) GetAllSession(ctx context.Context) ([]*dto.GetAllSessionResponse, error){

	sessions, err := c.chatSessionRepository.GetAllSession(ctx)
	if err != nil {
		return nil, err
	}

	response := make([]*dto.GetAllSessionResponse, 0)
	for _, sessions := range sessions {

		response = append(response, &dto.GetAllSessionResponse{
			Id: sessions.Id,
			Title: sessions.Title,
			CreateAt: sessions.CreateAt,
			UpdatedAt: sessions.UpdatedAt,
		})
		
	}

	return response, nil
}
