package service

import (
	"ai-notetaking-be/internal/constant"
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/repository"
	"ai-notetaking-be/pkg/chatbot"
	"ai-notetaking-be/pkg/embedding"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IChatbotService interface {
	CreateSession(ctx context.Context) (*dto.CreateSessionResponse, error)
	GetAllSession(ctx context.Context) ([]*dto.GetAllSessionResponse, error)
	GetChatHistory(ctx context.Context, sessionId uuid.UUID) ([]*dto.GetChatHistoryResponse, error)
	SendChat(ctx context.Context, request *dto.SendChatRequest) (*dto.SendChatResponse, error)
	DeleteSession(ctx context.Context, sessionId *dto.DeleteSessionRequest) error
}

type chatbotService struct {
	db                       *pgxpool.Pool
	chatSessionRepository    repository.IChatSessionRepository
	chatMessageRepository    repository.IChatMessageRepository
	chatMessageRawRepository repository.IChatMessageRawRepository
	notEmbeddingRepository   repository.INoteEmbeddingRepository
}

func NewChatbotService(
	db *pgxpool.Pool,
	chatSessionRepository repository.IChatSessionRepository,
	chatMessageRepository repository.IChatMessageRepository,
	chatMessageRawRepository repository.IChatMessageRawRepository,
	notEmbeddingRepository repository.INoteEmbeddingRepository,
) IChatbotService {
	return &chatbotService{
		db:                       db,
		chatSessionRepository:    chatSessionRepository,
		chatMessageRepository:    chatMessageRepository,
		chatMessageRawRepository: chatMessageRawRepository,
		notEmbeddingRepository:   notEmbeddingRepository,
	}
}

func (c *chatbotService) CreateSession(ctx context.Context) (*dto.CreateSessionResponse, error) {

	now := time.Now()
	chatSession := &entity.ChatSession{
		Id:       uuid.New(),
		Title:    "Unamed session",
		CreateAt: now,
	}

	chatMessage := &entity.ChatMessage{
		Id:            uuid.New(),
		Chat:          "Hi, how can i help you ?",
		Role:          constant.ChatMessageRoleModel,
		ChatSessionId: chatSession.Id,
		CreateAt:      now,
	}

	chatMessageRawUser := &entity.ChatMessageRaw{
		Id:            uuid.New(),
		Chat:          constant.ChatMessageRawInititalUserPromptV1,
		Role:          constant.ChatMessageRoleUser,
		ChatSessionId: chatSession.Id,
		CreateAt:      now,
	}

	chatMessageRawModel := &entity.ChatMessageRaw{
		Id:            uuid.New(),
		Chat:          constant.ChatMessageRawInititalModelPromptV1,
		Role:          constant.ChatMessageRoleModel,
		ChatSessionId: chatSession.Id,
		CreateAt:      now,
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

func (c *chatbotService) GetAllSession(ctx context.Context) ([]*dto.GetAllSessionResponse, error) {

	sessions, err := c.chatSessionRepository.GetAllSession(ctx)
	if err != nil {
		return nil, err
	}

	response := make([]*dto.GetAllSessionResponse, 0)
	for _, sessions := range sessions {

		response = append(response, &dto.GetAllSessionResponse{
			Id:        sessions.Id,
			Name:      sessions.Title,
			CreateAt:  sessions.CreateAt,
			UpdatedAt: sessions.UpdatedAt,
		})

	}

	return response, nil
}

func (c *chatbotService) GetChatHistory(ctx context.Context, sessionId uuid.UUID) ([]*dto.GetChatHistoryResponse, error) {

	_, err := c.chatSessionRepository.GetSessionById(ctx, sessionId)
	if err != nil {
		return nil, err
	}

	messages, err := c.chatSessionRepository.GetChatBySessionId(ctx, sessionId)
	if err != nil {
		return nil, err
	}

	response := make([]*dto.GetChatHistoryResponse, 0)
	for _, sessions := range messages {

		response = append(response, &dto.GetChatHistoryResponse{
			Id:        sessions.Id,
			Role:      sessions.Role,
			Chat:      sessions.Chat,
			CreatedAt: sessions.CreateAt,
		})

	}

	return response, nil
}

func (c *chatbotService) SendChat(ctx context.Context, request *dto.SendChatRequest) (*dto.SendChatResponse, error) {

	tx, err := c.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	chatSessionRepository := c.chatSessionRepository.UsingTx(ctx, tx)
	chatMessageRepository := c.chatMessageRepository.UsingTx(ctx, tx)
	chatMessageRawRepository := c.chatMessageRawRepository.UsingTx(ctx, tx)
	noteEmbeddingRepository := c.notEmbeddingRepository.UsingTx(ctx, tx)

	SessionChat, err := chatSessionRepository.GetSessionById(ctx, request.ChatSessionId)
	if err != nil {
		return nil, err
	}

	ExistingChatRaw, err := chatMessageRawRepository.GetChatBySessionId(ctx, request.ChatSessionId)
	if err != nil {
		return nil, err
	}

	updateSessionTitle := len(ExistingChatRaw) == 2
	now := time.Now()

	chatMessageUser := entity.ChatMessage{
		Id:            uuid.New(),
		Chat:          request.Chat,
		Role:          constant.ChatMessageRoleUser,
		ChatSessionId: request.ChatSessionId,
		CreateAt:      now,
	}

	embeddingRes, err := embedding.GetGeminiEmbedding(
		os.Getenv("GOOGLE_GEMINI_API_KEY"),
		request.Chat,
		"RETRIEVAL_QUERY",
	)

	if err != nil {
		return nil, err
	}

	decideUseRAGChatHistories := make([]*chatbot.ChatHistory, 0)
	for i, rawChat := range ExistingChatRaw {
		if i == 0 {
			decideUseRAGChatHistories = append(decideUseRAGChatHistories, &chatbot.ChatHistory{
				Chat: constant.DecideUseRAGMessageRawInitialUserPromptV1,
				Role: constant.ChatMessageRoleUser,
			})
		} else if i == 1 {
			decideUseRAGChatHistories = append(decideUseRAGChatHistories, &chatbot.ChatHistory{
				Chat: constant.ChatMessageRawInititalModelPromptV1,
				Role: constant.ChatMessageRoleModel,
			})
		}

		decideUseRAGChatHistories = append(decideUseRAGChatHistories, &chatbot.ChatHistory{
			Chat: rawChat.Chat,
			Role: rawChat.Role,
		})
	}

	useRAG, err := chatbot.DecideToUseRAG(
		ctx,
		os.Getenv("GOOGLE_GEMINI_API_KEY"),
		decideUseRAGChatHistories,
	)
	if err != nil {
		return nil, err
	}

	strBuilder := strings.Builder{}

	if useRAG {

		noteEmbeddings, err := noteEmbeddingRepository.SearchSimilarity(ctx, embeddingRes.Embedding.Values)
		if err != nil {
			return nil, err
		}

		for i, noteEmbeding := range noteEmbeddings {
			strBuilder.WriteString(fmt.Sprintf("Reference %d\n", i+1))
			strBuilder.WriteString(noteEmbeding.Document)
			strBuilder.WriteString("\n\n")
		}

	}

	strBuilder.WriteString("User Next Question: ")
	strBuilder.WriteString(request.Chat)
	strBuilder.WriteString("\n\n")
	strBuilder.WriteString("Your Answer")

	chatMessageRawUser := entity.ChatMessageRaw{
		Id:            uuid.New(),
		Chat:          strBuilder.String(),
		Role:          constant.ChatMessageRoleUser,
		ChatSessionId: request.ChatSessionId,
		CreateAt:      now,
	}

	ExistingChatRaw = append(
		ExistingChatRaw,
		&chatMessageRawUser,
	)

	geminiReq := make([]*chatbot.ChatHistory, 0)
	for _, ExistingChat := range ExistingChatRaw {

		geminiReq = append(geminiReq, &chatbot.ChatHistory{
			Chat: ExistingChat.Chat,
			Role: ExistingChat.Role,
		})
	}

	reply, err := chatbot.GetGeminiResponse(
		ctx,
		os.Getenv("GOOGLE_GEMINI_API_KEY"),
		geminiReq,
	)
	if err != nil {
		return nil, err
	}

	chatMessageModel := entity.ChatMessage{
		Id:            uuid.New(),
		Chat:          reply,
		Role:          constant.ChatMessageRoleModel,
		ChatSessionId: request.ChatSessionId,
		CreateAt:      now.Add(1 * time.Millisecond),
	}

	chatMessageRawModel := entity.ChatMessageRaw{
		Id:            uuid.New(),
		Chat:          reply,
		Role:          constant.ChatMessageRoleModel,
		ChatSessionId: request.ChatSessionId,
		CreateAt:      now.Add(1 * time.Millisecond),
	}

	chatMessageRepository.Create(ctx, &chatMessageUser)
	chatMessageRepository.Create(ctx, &chatMessageModel)
	chatMessageRawRepository.Create(ctx, &chatMessageRawUser)
	chatMessageRawRepository.Create(ctx, &chatMessageRawModel)

	if updateSessionTitle {
		SessionChat.Title = request.Chat
		SessionChat.UpdatedAt = &now
	}

	err = chatSessionRepository.Update(ctx, SessionChat)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.SendChatResponse{
		ChatSessionId: SessionChat.Id,
		Title:         SessionChat.Title,
		Send: &dto.SendChatResponseChat{
			Id:        chatMessageUser.Id,
			Chat:      chatMessageUser.Chat,
			Role:      chatMessageUser.Role,
			CreatedAt: chatMessageUser.CreateAt,
		},
		Reply: &dto.SendChatResponseChat{
			Id:        chatMessageModel.Id,
			Chat:      chatMessageModel.Chat,
			Role:      chatMessageModel.Role,
			CreatedAt: chatMessageModel.CreateAt,
		},
	}, nil
}

func (c *chatbotService) DeleteSession(ctx context.Context, session *dto.DeleteSessionRequest) error {

	tx, err := c.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	chatSessionRepository := c.chatSessionRepository.UsingTx(ctx, tx)
	chatMessageRepository := c.chatMessageRepository.UsingTx(ctx, tx)
	chatMessageRawRepository := c.chatMessageRawRepository.UsingTx(ctx, tx)

	_, err = chatSessionRepository.GetSessionById(ctx, session.ChatSessionId)
	if err != nil {
		return err
	}

	err = chatSessionRepository.Delete(ctx, session.ChatSessionId)
	if err != nil {
		return err
	}

	err = chatMessageRawRepository.DeleteBySessionId(ctx, session.ChatSessionId)
	if err != nil {
		return err
	}

	err = chatMessageRepository.DeleteBySessionId(ctx, session.ChatSessionId)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
