package service

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/internal/repository"
	"ai-notetaking-be/pkg/chunking"
	"ai-notetaking-be/pkg/embedding"
	garagestorages3 "ai-notetaking-be/pkg/garage-storage-s3"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tmc/langchaingo/schema"
)

type IConsumerService interface {
	Consume(ctx context.Context) error
}

type consumerService struct {
	notebookRepository      repository.INotebookRepository
	noteRepository          repository.INoteRepository
	noteEmbeddingRepository repository.INoteEmbeddingRepository
	pubSub                  *gochannel.GoChannel
	fileRepository          repository.IFileRepository
	s3Client                *garagestorages3.GarageS3
	topicName               string

	db *pgxpool.Pool
}

func (cs *consumerService) Consume(ctx context.Context) error {
	message, err := cs.pubSub.Subscribe(ctx, cs.topicName)
	if err != nil {
		return err
	}

	go func() {
		for msg := range message {
			cs.processMessage(ctx, msg)
		}
	}()

	return nil

}

func (cs *consumerService) processMessage(ctx context.Context, msg *message.Message) error {
	defer msg.Nack()

	defer func() {
		if e := recover(); e != nil {
			log.Errorf("[Panic Recovery] Terjadi panic saat memproses embedding: %v", e)
		}
	}()

	// =========================
	// Parse Payload
	// =========================
	var payload dto.PublishEmbedNoteMessage
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Errorf("[Consumer] Gagal unmarshal payload: %v | Payload: %s", err, string(msg.Payload))
		return err
	}

	// =========================
	// Ambil Note & Notebook
	// =========================
	note, err := cs.noteRepository.GetById(ctx, payload.NotedId)
	if err != nil {
		log.Errorf("[Repo] Gagal ambil note (ID: %s): %v", payload.NotedId, err)
		return err
	}

	notebook, err := cs.notebookRepository.GetById(ctx, note.NotebookId)
	if err != nil {
		log.Errorf("[Repo] Gagal ambil notebook (ID: %s) untuk note %s: %v", note.NotebookId, note.Id, err)
		return err
	}

	// =========================
	// File Metadata (Opsional)
	// =========================
	var (
		fileIDPtr    *uuid.UUID
		originalName string
	)

	fileMeta, err := cs.fileRepository.GetByNoteId(ctx, note.Id)
	if err != nil {
		log.Infof("[Repo] Note %s tidak memiliki lampiran file (opsional)", note.Id)
	} else if fileMeta != nil {
		fileIDPtr = &fileMeta.Id
		originalName = fileMeta.OriginalName
	}

	// =========================
	// Prepare Note Content (Page 0)
	// =========================
	noteUpdatedAt := "-"
	if note.UpdatedAt != nil {
		noteUpdatedAt = note.UpdatedAt.Format(time.RFC3339)
	}

	noteContent := fmt.Sprintf(`
Note Title      : %s
Notebook Title  : %s
File Referensi  : %s

%s

Created At      : %s
Updated At      : %s
`,
		note.Title,
		notebook.Name,
		originalName,
		note.Content,
		note.CreatedAt.Format(time.RFC3339),
		noteUpdatedAt,
	)

	// =========================
	// Chunking (HANYA via ChunkPdfPage)
	// =========================
	const maxChunkSize = 800
	var docs []schema.Document

	// NOTE → Page 0
	docs = append(docs,
		chunking.ChunkPdfPage(
			chunking.PdfPage{
				PageNumber: 0,
				Content:    noteContent,
			},
			maxChunkSize,
		)...,
	)

	// PDF → Per Page
	if fileMeta != nil {
		body, err := cs.s3Client.Download(ctx, fileMeta.Bucket, fileMeta.FileName)
		if err != nil {
			log.Errorf("[Storage] Download gagal: %v", err)
			return err
		}
		defer body.Close()

		pages, err := serverutils.ExtractTextPerPage(body)
		if err != nil {
			log.Errorf("[PDF] Extract text gagal: %v", err)
			return err
		}

		for _, page := range pages {
			docs = append(docs,
				chunking.ChunkPdfPage(
					chunking.PdfPage{
						PageNumber: page.PageNumber,
						Content:    page.Content,
					},
					maxChunkSize,
				)...,
			)
		}
	}

	// =========================
	// DB Transaction
	// =========================
	tx, err := cs.db.Begin(ctx)
	if err != nil {
		log.Errorf("[DB] Gagal memulai transaksi: %v", err)
		return err
	}
	defer tx.Rollback(ctx)

	repo := cs.noteEmbeddingRepository.UsingTx(ctx, tx)

	if err := repo.DeleteByID(ctx, note.Id); err != nil {
		log.Errorf("[DB] Gagal hapus embedding lama untuk note %s: %v", note.Id, err)
		return err
	}

	// =========================
	// Embedding Loop
	// =========================
	for i, doc := range docs {
		pageNumber := 0
		if v, ok := doc.Metadata["page"].(int); ok {
			pageNumber = v
		}

		log.Debugf(
			"[AI] Mengirim chunk %d/%d | Note:%s | Page:%d",
			i+1,
			len(docs),
			note.Id,
			pageNumber,
		)

		res, err := embedding.GetGeminiEmbedding(
			os.Getenv("GOOGLE_GEMINI_API_KEY"),
			"models/gemini-embedding-exp-03-07",
			doc.PageContent,
			"RETRIEVAL_DOCUMENT",
		)
		if err != nil {
			log.Errorf("[AI] Gagal mendapatkan embedding dari Gemini: %v", err)
			return err
		}

		if err := repo.Create(ctx, &entity.NoteEmbedding{
			Id:             uuid.New(),
			NoteId:         note.Id,
			FileId:         fileIDPtr,
			ChunkContent:   doc.PageContent,
			EmbeddingValue: res.Embedding.Values,
			PageNumber:     pageNumber,
			ChunkIndex:     i + 1,
			OverlapRange:   "none",
			CreatedAt:      time.Now(),
		}); err != nil {
			log.Errorf(
				"[DB] Gagal simpan embedding Note %s | Chunk %d: %v",
				note.Id,
				i+1,
				err,
			)
			return err
		}
	}

	// =========================
	// Commit
	// =========================
	if err := tx.Commit(ctx); err != nil {
		log.Errorf("[DB] Gagal commit transaksi: %v", err)
		return err
	}

	log.Infof(
		"[Success] Berhasil memproses %d chunk embedding untuk Note: %s",
		len(docs),
		note.Id,
	)

	msg.Ack()
	return nil
}

func NewConsumerService(
	pubSub *gochannel.GoChannel,
	topicName string,
	noteRepository repository.INoteRepository,
	noteEmbeddingRepository repository.INoteEmbeddingRepository,
	notebookRepository repository.INotebookRepository,
	fileRepository repository.IFileRepository,
	s3Client *garagestorages3.GarageS3,
	db *pgxpool.Pool) IConsumerService {
	return &consumerService{
		pubSub:                  pubSub,
		topicName:               topicName,
		noteRepository:          noteRepository,
		noteEmbeddingRepository: noteEmbeddingRepository,
		notebookRepository:      notebookRepository,
		fileRepository:          fileRepository,
		s3Client:                s3Client,
		db:                      db,
	}
}
