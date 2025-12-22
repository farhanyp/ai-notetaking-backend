package service

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/gofiber/fiber/v2/log"
)

type EmbeddingRequestContentPart struct{
	Text string `json:"text"`
}

type EmbeddingRequestContent struct{
	Parts []EmbeddingRequestContentPart `json:"parts"`
}

type EmbeddingRequest struct {
	Model string `json:"model"`
	Content EmbeddingRequestContent `json:"content"`
	TaskType string `json:"task_type"`
}

type IConsumerService interface {
	Consume(ctx context.Context) error
}

type consumerService struct {
	noteRepository repository.INoteRepository
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

	geminiReq := EmbeddingRequest{
		Model: "models/gemini-embedding-exp-03-07",
		Content: EmbeddingRequestContent{
			Parts: []EmbeddingRequestContentPart{
				{
					Text: note.Content,
				},
			},
		},
		TaskType: "SEMANTIC_SIMILARITY",
	}

	http.NewRequest(
		"Post",
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-001:embedContent",
		geminiReq,
	)

	client := &http.Client{}
	client.Do()

	fmt.Printf("Processing note ID: %s\n", payload.NotedId)

	msg.Ack()

	return nil

}

func NewConsumerService(pubSub *gochannel.GoChannel, topicName string, noteRepository repository.INoteRepository) IConsumerService {
	return &consumerService{
		pubSub: pubSub,
		topicName: topicName,
		noteRepository: noteRepository,
	}
}