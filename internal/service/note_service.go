package service

import (
	"ai-notetaking-be/internal/constant"
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/internal/repository"
	"ai-notetaking-be/pkg/chatbot"
	"ai-notetaking-be/pkg/embedding"
	garagestorages3 "ai-notetaking-be/pkg/garage-storage-s3"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
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
	ExtractPreview(ctx context.Context, noteId uuid.UUID) (string, error)
	ExtractPreviewWithAI(ctx context.Context, noteId uuid.UUID) (string, error)
	UpdateFromExtraction(ctx context.Context, req *dto.UpdateNoteRequest) (*dto.UpdateNoteResponse, error)
}

type noteService struct {
	noteRepository         repository.INoteRepository
	notebookRepository     repository.INotebookRepository
	fileRepository         repository.IFileRepository
	s3Client               *garagestorages3.GarageS3
	publisherService       IPublisherService
	notEmbeddingRepository repository.INoteEmbeddingRepository
	db                     *pgxpool.Pool
}

func NewNoteService(
	noteRepository repository.INoteRepository,
	notebookRepository repository.INotebookRepository,
	fileRepository repository.IFileRepository,
	s3Client *garagestorages3.GarageS3,
	publisherService IPublisherService,
	notEmbeddingRepository repository.INoteEmbeddingRepository,
	db *pgxpool.Pool,
) INoteService {
	return &noteService{
		noteRepository:         noteRepository,
		notebookRepository:     notebookRepository,
		fileRepository:         fileRepository,
		s3Client:               s3Client,
		publisherService:       publisherService,
		notEmbeddingRepository: notEmbeddingRepository,
		db:                     db,
	}
}

func (c *noteService) Create(ctx context.Context, req *dto.CreateNoteRequest) (*dto.CreateNoteResponse, error) {

	note := entity.Note{
		Id:         uuid.New(),
		Title:      req.Title,
		Content:    req.Content,
		NotebookId: req.NotebookId,
		CreatedAt:  time.Now(),
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
		Id:         note.Id,
		Title:      note.Title,
		NotebookId: note.NotebookId,
		Content:    note.Content,
		CreatedAt:  note.CreatedAt,
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
					NotebookId: noteItem.NotebookId,
					CreatedAt:  noteItem.CreatedAt,
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
	fileRepository := c.fileRepository.UsingTx(ctx, tx)

	err = noteRepository.DeleteById(ctx, idParam)
	if err != nil {
		return err
	}

	err = noteEmbeddingRepository.DeleteByID(ctx, idParam)
	if err != nil {
		return err
	}

	err = fileRepository.DeleteByNoteId(ctx, idParam)
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

	if req.NotebookId != nil {
		_, err = c.notebookRepository.GetById(ctx, *req.NotebookId)
		if err != nil {
			return nil, err
		}
	}

	now := time.Now()
	note.NotebookId = *req.NotebookId
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

func (s *noteService) ExtractPreview(ctx context.Context, noteId uuid.UUID) (string, error) {

	note, err := s.noteRepository.GetById(ctx, noteId)
	if err != nil {
		return "", err
	}

	var preview strings.Builder

	// content manual dari note (kalau ada)
	if strings.TrimSpace(note.Content) != "" {
		preview.WriteString(note.Content)
		preview.WriteString("\n\n")
	}

	fileMeta, err := s.fileRepository.GetByNoteId(ctx, note.Id)
	if err != nil || fileMeta == nil {
		// langsung return string mentah
		return preview.String(), nil
	}

	body, err := s.s3Client.Download(ctx, fileMeta.Bucket, fileMeta.FileName)
	if err != nil {
		return "", err
	}
	defer body.Close()

	pages, err := serverutils.ExtractTextPerPage(body)
	if err != nil {
		return "", err
	}

	for _, page := range pages {
		if strings.TrimSpace(page.Content) == "" {
			continue
		}

		preview.WriteString(page.Content)
		preview.WriteString("\n\n")
	}

	return preview.String(), nil
}

func (s *noteService) ExtractPreviewWithAI(ctx context.Context, noteId uuid.UUID) (string, error) {
	// 1. Ambil data note dari database
	note, err := s.noteRepository.GetById(ctx, noteId)
	if err != nil {
		return "", err
	}

	// 2. Ambil metadata file yang diasosiasikan dengan note
	fileMeta, err := s.fileRepository.GetByNoteId(ctx, note.Id)
	if err != nil || fileMeta == nil {
		return "", fmt.Errorf("no file associated with this note")
	}

	// 3. Download file dari S3
	body, err := s.s3Client.Download(ctx, fileMeta.Bucket, fileMeta.FileName)
	if err != nil {
		return "", err
	}
	defer body.Close()

	// 4. Ekstrak teks per halaman dari PDF
	pages, err := serverutils.ExtractTextPerPage(body)
	if err != nil {
		return "", err
	}

	// 5. Gabungkan konten dari semua halaman ke dalam strings.Builder
	var rawText strings.Builder
	for _, page := range pages {
		if strings.TrimSpace(page.Content) == "" {
			continue
		}

		rawText.WriteString(page.Content)
		rawText.WriteString("\n\n")
	}

	// Validasi jika hasil ekstraksi kosong
	if rawText.Len() == 0 {
		return "", fmt.Errorf("extracted pdf text is empty")
	}

	// 6. Konversi strings.Builder menjadi string tunggal untuk dikirim ke AI
	content := rawText.String()

	// 7. Siapkan request untuk Gemini
	geminiReq := make([]*chatbot.ChatHistory, 0)

	// Memberikan konteks sistem (System Instruction)
	geminiReq = append(geminiReq, &chatbot.ChatHistory{
		Role: constant.ChatMessageRoleModel,
		Chat: "As an expert in markdown formatting for ReactMarkdown, your task is to transform an extracted text into a well-formatted markdown document for ReactMarkdown.",
	})

	// Memberikan instruksi user dan menyuntikkan konten teks
	geminiReq = append(geminiReq, &chatbot.ChatHistory{
		Role: constant.ChatMessageRoleUser,
		Chat: fmt.Sprintf(`
        Instructions:
        1  -  Input:
        You will receive plain text extracted from function ExtractTextFromPdf use library github.com/ledongthuc/pdf.

        2  -  Output:
        Provide the same text but formatted in markdown react, don't change anything, don't add uppercase, don't add new line, the output should be Markdown with the same input text as it is.

        3  -  Formatting Requirements:
        *Headers: Identify and convert headings to markdown headers (e.g., # Header 1, ## Header 2, etc.).
        *Lists: Detect and format lists (both ordered and unordered).
        *Emphasis: Apply appropriate emphasis using *italic*, **bold**, or ***bold italic*** where needed.
        *Links: Convert URLs into markdown link format [link text](URL).
        *Code Blocks: Format any code snippets as inline code or code blocks.
        *Blockquotes: Apply blockquote formatting to any quoted text.
        *Tables: Convert any tabular data into markdown tables.

        ---
        INPUT TEXT: %s

        OUTPUT ONLY THE MARKDOWN TEXT, DON'T OUTPUT ANYTHING ELSE.

        NOW CONVERT TEXT.`, content),
	})

	// 8. Panggil API Gemini
	reply, err := chatbot.GetGeminiResponse(
		ctx,
		os.Getenv("GOOGLE_GEMINI_API_KEY"),
		geminiReq,
	)
	if err != nil {
		log.Printf("[ExtractPreviewWithAI] Gemini error for note %s: %v", note.Id, err)
		return "", err
	}

	// 9. Kembalikan hasil yang sudah dinormalisasi
	return reply, nil
}

func (s *noteService) UpdateFromExtraction(ctx context.Context, req *dto.UpdateNoteRequest) (*dto.UpdateNoteResponse, error) {
	// 1. Ambil data note lama
	note, err := s.noteRepository.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// 2. Update konten dengan teks yang sudah di-approve/edit user
	now := time.Now()
	note.Content = req.Content
	note.UpdatedAt = &now

	err = s.noteRepository.Update(ctx, note)
	if err != nil {
		return nil, err
	}

	// 3. Trigger Re-Indexing via Publisher
	payload := dto.PublishEmbedNoteMessage{
		NotedId: note.Id,
	}
	payloadJson, _ := json.Marshal(payload)
	_ = s.publisherService.Publish(ctx, payloadJson)

	return &dto.UpdateNoteResponse{Id: note.Id}, nil
}
