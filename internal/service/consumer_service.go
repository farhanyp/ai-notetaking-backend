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

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

type IConsumerService interface {
	Consume(ctx context.Context) error
}

type consumerService struct {
	noteRepository repository.INoteRepository
	noteEmbeddingRepository repository.INoteEmbeddingRepository
	pubSub *gochannel.GoChannel
	topicName string
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
			log.Error(e)
		}
	}()

	var payload dto.PublishEmbedNoteMessage
	err := json.Unmarshal(msg.Payload, &payload)
	if err != nil {
		panic(err)
	}

	note, err := cs.noteRepository.GetById(ctx, payload.NotedId)
	if err != nil {
		panic(err)
	}

	res, err := embedding.GetGeminiEmbedding(
		os.Getenv("GOOGLE_GEMINI_API_KEY"),
		note.Content,
	)
	if  err != nil {
		panic(err)
	}

	noteEmbedding := entity.NoteEmbedding{
		Id: uuid.New(),
		Document: note.Content,
		EmbeddingValue: res.Embedding.Values,
		NoteId: note.Id,
		CreateAt: time.Now(),
	}

	err = cs.noteEmbeddingRepository.Create(ctx, &noteEmbedding)
	if  err != nil {
		panic(err)
	}

	msg.Ack()

	return nil

}

func NewConsumerService(pubSub *gochannel.GoChannel, topicName string, noteRepository repository.INoteRepository, noteEmbeddingRepository repository.INoteEmbeddingRepository) IConsumerService {
	return &consumerService{
		pubSub: pubSub,
		topicName: topicName,
		noteRepository: noteRepository,
		noteEmbeddingRepository: noteEmbeddingRepository,
	}
}