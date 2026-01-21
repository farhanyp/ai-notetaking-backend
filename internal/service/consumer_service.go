package service

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/repository"
	"ai-notetaking-be/pkg/chunking"
	"ai-notetaking-be/pkg/embedding"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

	var payload dto.PublishEmbedNoteMessage
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Errorf("[Consumer] Gagal unmarshal payload: %v | Payload: %s", err, string(msg.Payload))
		return err
	}

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

	fileMeta, err := cs.fileRepository.GetByNoteId(ctx, note.Id)
	var fileIDPtr *uuid.UUID
	if err != nil {
		log.Infof("[Repo] Note %s tidak memiliki lampiran file (Opsional)", note.Id)
	} else if fileMeta != nil {
		fileIDPtr = &fileMeta.Id
	}

	noteUpdatedAt := "-"
	if note.UpdatedAt != nil {
		noteUpdatedAt = note.UpdatedAt.Format(time.RFC3339)
	}

	content := fmt.Sprintf(`
    Note Title : %s
    Notebook Title: %s
	File referensi: %s

    %s

    Created at: %s
    Updated at: %s
    `,
		note.Title,
		notebook.Name,
		fileMeta.OriginalName,
		note.Content,
		note.CreatedAt.Format(time.RFC3339),
		noteUpdatedAt,
	)

	var docs []schema.Document
	chunkSize := 500
	overlapPercentage := 10

	if len(content) > chunkSize {
		docs, err = chunking.SplitText(ctx, strings.NewReader(content), chunkSize, overlapPercentage)
		if err != nil {
			log.Errorf("[Splitter] Gagal melakukan chunking: %v", err)
			return err
		}
	} else {
		docs = []schema.Document{{PageContent: content}}
	}

	tx, err := cs.db.Begin(ctx)
	if err != nil {
		log.Errorf("[DB] Gagal memulai transaksi: %v", err)
		return err
	}
	defer tx.Rollback(ctx)

	noteEmbeddingRepository := cs.noteEmbeddingRepository.UsingTx(ctx, tx)

	if err := noteEmbeddingRepository.DeleteByID(ctx, note.Id); err != nil {
		log.Errorf("[DB] Gagal hapus embedding lama untuk note %s: %v", note.Id, err)
		return err
	}

	for i, doc := range docs {
		log.Debugf("[AI] Mengirim chunk %d/%d ke Gemini Embedding untuk Note: %s", i+1, len(docs), note.Id)
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

		noteEmbedding := entity.NoteEmbedding{
			Id:             uuid.New(),
			NoteId:         note.Id,
			FileId:         fileIDPtr,
			ChunkContent:   doc.PageContent,
			EmbeddingValue: res.Embedding.Values,
			PageNumber:     1, // Belum ada logika untuk melakukan page number
			ChunkIndex:     i + 1,
			OverlapRange:   fmt.Sprintf("%d%%", overlapPercentage),
			CreatedAt:      time.Now(),
		}

		if err := noteEmbeddingRepository.Create(ctx, &noteEmbedding); err != nil {
			log.Errorf("[DB] Gagal simpan embedding baru untuk note %s chunk %d: %v", note.Id, i+1, err)
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Errorf("[DB] Gagal commit transaksi: %v", err)
		return err
	}

	log.Infof("[Success] Berhasil memproses %d chunk embedding untuk Note: %s", len(docs), note.Id)
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
	db *pgxpool.Pool) IConsumerService {
	return &consumerService{
		pubSub:                  pubSub,
		topicName:               topicName,
		noteRepository:          noteRepository,
		noteEmbeddingRepository: noteEmbeddingRepository,
		notebookRepository:      notebookRepository,
		fileRepository:          fileRepository,
		db:                      db,
	}
}
