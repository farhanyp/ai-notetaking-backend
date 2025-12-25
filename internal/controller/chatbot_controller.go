package controller

import (
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/internal/service"

	"github.com/gofiber/fiber/v2"
)

type IChatbotController interface {
	RegisterRoutes(r fiber.Router)
	CreateSession(ctx *fiber.Ctx) error
	GetAllSession(ctx *fiber.Ctx) error
}

type chatbotController struct {
	chatbotService service.IChatbotService
}

func NewChatController(chatbotService service.IChatbotService) IChatbotController {
	return &chatbotController{
		chatbotService: chatbotService,
	}
}

func (c *chatbotController) RegisterRoutes(r fiber.Router) {
	h := r.Group("/v1/chatbot")
	h.Post("/create-session", c.CreateSession)
	h.Get("/create-session", c.GetAllSession)
}

func (c *chatbotController) CreateSession(ctx *fiber.Ctx) error {

	res, err := c.chatbotService.CreateSession(ctx.Context())
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success create session", res))
}

func (c *chatbotController) GetAllSession(ctx *fiber.Ctx) error {

	res, err := c.chatbotService.GetAllSession(ctx.Context())
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success get all session", res))
}
