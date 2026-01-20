package service

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/repository"
	garagestorages3 "ai-notetaking-be/pkg/garage-storage-s3"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type INotebookService interface {
	GetAll(ctx context.Context) ([]*dto.ListNotebookResponse, error)
	Create(ctx context.Context, req *dto.CreateNotebookRequest) (*dto.CreateNotebookResponse, error)
	Show(ctx context.Context, id uuid.UUID) (*dto.ShowNotebookResponse, error)
	Update(ctx context.Context, req *dto.UpdateNotebookRequest) (*dto.UpdateNotebookResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Move(ctx context.Context, req *dto.MoveNotebookRequest) (*dto.MoveNotebookResponse, error)
}

type notebookService struct {
	notebookRepository      repository.INotebookRepository
	noteRepository          repository.INoteRepository
	noteEmbeddingRepository repository.INoteEmbeddingRepository
	publisherService        IPublisherService
	fileRepository          repository.IFileRepository
	s3Client                *garagestorages3.GarageS3

	db *pgxpool.Pool
}

func NewNotebookService(
	notebookRepository repository.INotebookRepository,
	noteRepository repository.INoteRepository,
	noteEmbeddingRepository repository.INoteEmbeddingRepository,
	publisherService IPublisherService,
	fileRepository repository.IFileRepository,
	s3Client *garagestorages3.GarageS3,
	db *pgxpool.Pool) INotebookService {
	return &notebookService{
		notebookRepository:      notebookRepository,
		noteRepository:          noteRepository,
		noteEmbeddingRepository: noteEmbeddingRepository,
		publisherService:        publisherService,
		db:                      db,
		fileRepository:          fileRepository,
		s3Client:                s3Client,
	}
}

func (c *notebookService) GetAll(ctx context.Context) ([]*dto.ListNotebookResponse, error) {
	// 1. Ambil semua Notebooks
	notebooks, err := c.notebookRepository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	notebookIds := make([]uuid.UUID, 0)
	result := make([]*dto.ListNotebookResponse, 0)
	for _, notebook := range notebooks {
		res := dto.ListNotebookResponse{
			Id:        notebook.Id,
			Name:      notebook.Name,
			ParentId:  notebook.ParentId,
			CreatedAt: notebook.CreatedAt,
			UpdateAt:  notebook.UpdatedAt,
			Notes:     make([]*dto.GetAllNotebookResponseNote, 0),
		}
		result = append(result, &res)
		notebookIds = append(notebookIds, notebook.Id)
	}

	// 2. Ambil semua Notes berdasarkan Notebook IDs
	notes, err := c.noteRepository.GetByNotesIds(ctx, notebookIds)
	if err != nil {
		return nil, err
	}

	noteIds := make([]uuid.UUID, len(notes))
	for i, n := range notes {
		noteIds[i] = n.Id
	}

	// 3. Ambil semua Files dan Generate Presigned URL
	files, err := c.fileRepository.GetByNoteIds(ctx, noteIds)
	if err != nil {
		// Log error dari database
		fmt.Printf("[ERROR] Failed to fetch files from database: %v\n", err)
	}

	// Grouping files by NoteId: map[NoteId][]dto.NoteFileDTO
	fileMap := make(map[uuid.UUID][]dto.NoteFileDTO)

	if len(files) > 0 {
		fmt.Printf("[INFO] Processing %d files to generate presigned URLs\n", len(files))

		for _, f := range files {
			// Generate Presigned URL dari GarageS3 (berlaku misal 1 jam)
			url, err := c.s3Client.GetPresignedURL(ctx, f.Bucket, f.FileName, time.Hour*1)
			if err != nil {
				// Log detail file mana yang gagal di-generate URL-nya
				fmt.Printf("[ERROR] Failed to generate presigned URL for FileID: %s (Bucket: %s, Key: %s): %v\n",
					f.Id, f.Bucket, f.FileName, err)
				continue
			}

			fileMap[f.NoteId] = append(fileMap[f.NoteId], dto.NoteFileDTO{
				Name: f.OriginalName,
				Url:  url,
			})
		}
	} else {
		fmt.Println("[DEBUG] No files found for the given notes")
	}

	// 4. Gabungkan Data (O(n))
	for _, notebookRes := range result {
		for _, note := range notes {
			if note.NotebookId == notebookRes.Id {
				// Ambil list file dari map, jika tidak ada berikan array kosong
				attachedFiles := fileMap[note.Id]
				if attachedFiles == nil {
					attachedFiles = []dto.NoteFileDTO{}
				}

				notebookRes.Notes = append(notebookRes.Notes, &dto.GetAllNotebookResponseNote{
					Id:        note.Id,
					Title:     note.Title,
					Content:   note.Content,
					Files:     attachedFiles, // Masukkan array file
					CreatedAt: note.CreatedAt,
					UpdateAt:  note.UpdatedAt,
				})
			}
		}
	}

	return result, nil
}

func (c *notebookService) Create(ctx context.Context, req *dto.CreateNotebookRequest) (*dto.CreateNotebookResponse, error) {

	notebook := entity.Notebook{
		Id:        uuid.New(),
		Name:      req.Name,
		ParentId:  req.ParentId,
		CreatedAt: time.Now(),
	}

	err := c.notebookRepository.Create(ctx, &notebook)
	if err != nil {
		return nil, err
	}

	return &dto.CreateNotebookResponse{
		Id: notebook.Id,
	}, nil
}

func (c *notebookService) Update(ctx context.Context, req *dto.UpdateNotebookRequest) (*dto.UpdateNotebookResponse, error) {

	notebook, err := c.notebookRepository.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	notebook.Name = req.Name
	notebook.UpdatedAt = &now

	err = c.notebookRepository.Update(ctx, notebook)
	if err != nil {
		return nil, err
	}

	notes, err := c.noteRepository.GetByNotesIds(ctx, []uuid.UUID{notebook.Id})
	if err != nil {
		return nil, err
	}

	for _, note := range notes {
		msg := dto.PublishEmbedNoteMessage{
			NotedId: note.Id,
		}

		msgJson, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}

		err = c.publisherService.Publish(ctx, msgJson)
		if err != nil {
			return nil, err
		}

	}

	return &dto.UpdateNotebookResponse{
		Id: notebook.Id,
	}, nil
}

func (c *notebookService) Move(ctx context.Context, req *dto.MoveNotebookRequest) (*dto.MoveNotebookResponse, error) {

	_, err := c.notebookRepository.GetById(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	if req.ParentId != nil {
		_, err = c.notebookRepository.GetById(ctx, *req.ParentId)
		if err != nil {
			return nil, err
		}
	}

	err = c.notebookRepository.Move(ctx, req.Id, req.ParentId)
	if err != nil {
		return nil, err
	}

	return &dto.MoveNotebookResponse{
		Id: req.Id,
	}, nil
}

func (c *notebookService) Show(ctx context.Context, idParam uuid.UUID) (*dto.ShowNotebookResponse, error) {

	notebook, err := c.notebookRepository.GetById(ctx, idParam)

	if err != nil {
		return nil, err
	}

	res := dto.ShowNotebookResponse{
		Id:        notebook.Id,
		Name:      notebook.Name,
		ParentId:  notebook.ParentId,
		CreatedAt: notebook.CreatedAt,
	}

	return &res, nil
}

func (c *notebookService) Delete(ctx context.Context, idParam uuid.UUID) error {

	_, err := c.notebookRepository.GetById(ctx, idParam)
	if err != nil {
		return err
	}

	tx, err := c.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	notebookRepo := c.notebookRepository.UsingTx(ctx, tx)
	noteRepo := c.noteRepository.UsingTx(ctx, tx)
	noteEmbeddingRepo := c.noteEmbeddingRepository.UsingTx(ctx, tx)
	fileRepo := c.fileRepository.UsingTx(ctx, tx)

	err = notebookRepo.DeleteById(ctx, idParam)
	if err != nil {
		return err
	}

	err = noteEmbeddingRepo.DeleteByNotebookId(ctx, idParam)
	if err != nil {
		return err
	}

	err = notebookRepo.NullifyParentById(ctx, idParam)
	if err != nil {
		return err
	}

	err = noteRepo.DeleteByNotebookId(ctx, idParam)
	if err != nil {
		return err
	}

	err = fileRepo.DeleteByNotebookId(ctx, idParam)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
