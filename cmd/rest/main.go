package main

import (
	"ai-notetaking-be/internal/controller"
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/internal/repository"
	"ai-notetaking-be/internal/service"
	"ai-notetaking-be/pkg/database"
	garagestorages3 "ai-notetaking-be/pkg/garage-storage-s3"
	"context"
	"log"
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024,
	})

	app.Use(cors.New())
	app.Use(serverutils.ErrorHandlerMiddleware())

	s3Config := garagestorages3.Config{
		AccessKey: os.Getenv("GARAGE_S3_ACCESS_KEY"),
		SecretKey: os.Getenv("GARAGE_S3_SECRET_KEY"),
		Endpoint:  os.Getenv("BASE_URL"),
		Region:    os.Getenv("REGION"),
	}

	s3Client, err := garagestorages3.NewGarageClient(s3Config)
	if err != nil {
		panic(err)
	}

	db := database.ConnectDB(os.Getenv("DB_CONNECTION_STRING"))

	exampleRepository := repository.NewExampleRepository(db)
	fileRepository := repository.NewFileRepository(db)
	notebookRepository := repository.NewNotebookRepository(db)
	noteRepository := repository.NewNoteRepository(db)
	noteEmbeddingRepository := repository.NewNoteEmbeddingRepository(db)
	chatSessionRepository := repository.NewChatSessionRepository(db)
	chatMessageRepository := repository.NewChatMessageRepository(db)
	chatMessageRawRepository := repository.NewChatMessageRawRepository(db)

	watermillLogger := watermill.NewStdLogger(false, false)
	pubSub := gochannel.NewGoChannel(gochannel.Config{}, watermillLogger)
	publisherService := service.NewPublisherService(
		os.Getenv("EMBED_NOTE_CONTENT_TOPIC_NAME"),
		pubSub,
	)

	consumerService := service.NewConsumerService(
		pubSub,
		os.Getenv("EMBED_NOTE_CONTENT_TOPIC_NAME"),
		noteRepository,
		noteEmbeddingRepository,
		notebookRepository,
		db,
	)

	exampleService := service.NewExampleService(exampleRepository, s3Client)
	notebookService := service.NewNotebookService(notebookRepository, noteRepository, noteEmbeddingRepository, publisherService, db)
	noteService := service.NewNoteService(noteRepository, notebookRepository, fileRepository, s3Client, publisherService, noteEmbeddingRepository, db)
	chatbotService := service.NewChatbotService(db, chatSessionRepository, chatMessageRepository, chatMessageRawRepository, noteEmbeddingRepository)
	fileService := service.NewFileService(noteRepository, fileRepository, s3Client)

	exampleController := controller.NewExampleController(exampleService)
	notebookController := controller.NewNotebookController(notebookService)
	noteController := controller.NewNoteController(noteService)
	chatbotController := controller.NewChatController(chatbotService)
	fileController := controller.NewFileController(fileService)

	api := app.Group("/api")
	exampleController.RegisterRoutes(api)
	notebookController.RegisterRoutes(api)
	noteController.RegisterRoutes(api)
	chatbotController.RegisterRoutes(api)
	fileController.RegisterRoutes(api)

	err = consumerService.Consume(context.Background())
	if err != nil {
		panic(err)
	}

	log.Fatal(app.Listen(":3000"))
}
