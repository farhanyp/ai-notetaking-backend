package service

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/repository"
	"ai-notetaking-be/pkg/embedding"
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type INoteService interface {
	Create(ctx context.Context, req *dto.CreateNoteRequest) (*dto.CreateNoteResponse, error)
	Show(ctx context.Context, id uuid.UUID) (*dto.ShowNoteResponse, error)
	SemanticSearch(ctx context.Context, query string) ([]*dto.SemanticSearchResponse, error)
	Update(ctx context.Context, req *dto.UpdateNoteRequest) (*dto.UpdateNoteResponse, error)
	Delete(ctx context.Context, idParam uuid.UUID) error
	Move(ctx context.Context, req *dto.MoveNoteRequest) (*dto.MoveNoteResponse, error)
}

type noteService struct {
	noteRepository         repository.INoteRepository
	notebookRepository     repository.INotebookRepository
	publisherService       IPublisherService
	notEmbeddingRepository repository.INoteEmbeddingRepository
	db                     *pgxpool.Pool
}

func NewNoteService(
	noteRepository repository.INoteRepository,
	notebookRepository repository.INotebookRepository,
	publisherService IPublisherService,
	notEmbeddingRepository repository.INoteEmbeddingRepository,
	db *pgxpool.Pool,
) INoteService {
	return &noteService{
		noteRepository:         noteRepository,
		notebookRepository:     notebookRepository,
		publisherService:       publisherService,
		notEmbeddingRepository: notEmbeddingRepository,
		db:                     db,
	}
}

func (c *noteService) Create(ctx context.Context, req *dto.CreateNoteRequest) (*dto.CreateNoteResponse, error) {

	note := entity.Note{
		Id:          uuid.New(),
		Title:       req.Title,
		Content:     req.Content,
		Notebook_id: req.Notebook_id,
		CreateAt:    time.Now(),
	}

	err := c.noteRepository.Create(ctx, &note)
	if err != nil {
		return nil, err
	}

	msgPayload := dto.PublishEmbedNoteMessage{
		NotedId: note.Id,
	}

	msgJson, err := json.Marshal(msgPayload)
	if err != nil {
		return nil, err
	}

	err = c.publisherService.Publish(ctx, msgJson)
	if err != nil {
		return nil, err
	}

	return &dto.CreateNoteResponse{
		Id: note.Id,
	}, nil
}

func (c *noteService) Show(ctx context.Context, idParam uuid.UUID) (*dto.ShowNoteResponse, error) {

	note, err := c.noteRepository.GetById(ctx, idParam)

	if err != nil {
		return nil, err
	}

	res := dto.ShowNoteResponse{
		Id:          note.Id,
		Title:       note.Title,
		Notebook_id: note.Notebook_id,
		Content:     note.Content,
		CreateAt:    note.CreateAt,
	}

	return &res, nil
}

func (c *noteService) SemanticSearch(ctx context.Context, query string) ([]*dto.SemanticSearchResponse, error) {

	embeddingRes, err := embedding.GetGeminiEmbedding(
		os.Getenv("GOOGLE_GEMINI_API_KEY"),
		"models/gemini-embedding-exp-03-07",
		query,
		"RETRIEVAL_QUERY",
	)

	if err != nil {
		return nil, err
	}

	noteEmbeddings, err := c.notEmbeddingRepository.SemanticSearch(ctx, embeddingRes.Embedding.Values)
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0)
	for _, noteEmbedding := range noteEmbeddings {
		ids = append(ids, noteEmbedding.NoteId)
	}

	notes, err := c.noteRepository.GetByIds(ctx, ids)
	if err != nil {
		return nil, err
	}

	response := make([]*dto.SemanticSearchResponse, 0)
	for _, n := range noteEmbeddings {
		for _, noteItem := range notes {
			if n.NoteId == noteItem.Id {
				response = append(response, &dto.SemanticSearchResponse{
					Id:         noteItem.Id,
					Title:      noteItem.Title,
					Content:    noteItem.Content,
					NotebookId: noteItem.Notebook_id,
					CreateAt:   noteItem.CreateAt,
					UpdateAt:   noteItem.UpdatedAt,
				})
			}
		}
	}

	return response, nil
}

func (c *noteService) Update(ctx context.Context, req *dto.UpdateNoteRequest) (*dto.UpdateNoteResponse, error) {

	note, err := c.noteRepository.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	note.Title = req.Title
	note.Content = req.Content
	note.UpdatedAt = &now

	err = c.noteRepository.Update(ctx, note)
	if err != nil {
		return nil, err
	}

	payload := dto.PublishEmbedNoteMessage{
		NotedId: note.Id,
	}

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	err = c.publisherService.Publish(ctx, payloadJson)
	if err != nil {
		return nil, err
	}

	return &dto.UpdateNoteResponse{
		Id: note.Id,
	}, nil
}

func (c *noteService) Delete(ctx context.Context, idParam uuid.UUID) error {

	_, err := c.noteRepository.GetById(ctx, idParam)
	if err != nil {
		return err
	}

	tx, err := c.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	noteRepository := c.noteRepository.UsingTx(ctx, tx)

	noteEmbeddingRepository := c.notEmbeddingRepository.UsingTx(ctx, tx)

	err = noteRepository.DeleteById(ctx, idParam)
	if err != nil {
		return err
	}

	err = noteEmbeddingRepository.DeleteByID(ctx, idParam)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (c *noteService) Move(ctx context.Context, req *dto.MoveNoteRequest) (*dto.MoveNoteResponse, error) {

	note, err := c.noteRepository.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	if req.Notebook_id != nil {
		_, err = c.notebookRepository.GetById(ctx, *req.Notebook_id)
		if err != nil {
			return nil, err
		}
	}

	now := time.Now()
	note.Notebook_id = *req.Notebook_id
	note.UpdatedAt = &now

	err = c.noteRepository.Update(ctx, note)
	if err != nil {
		return nil, err
	}

	payload := dto.PublishEmbedNoteMessage{
		NotedId: note.Id,
	}

	payloadJson, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	err = c.publisherService.Publish(ctx, payloadJson)
	if err != nil {
		return nil, err
	}

	return &dto.MoveNoteResponse{
		Id: req.Id,
	}, nil
}
