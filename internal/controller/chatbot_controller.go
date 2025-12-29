package controller

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type IChatbotController interface {
	RegisterRoutes(r fiber.Router)
	CreateSession(ctx *fiber.Ctx) error
	GetAllSession(ctx *fiber.Ctx) error
	GetChatHistory(ctx *fiber.Ctx) error
	SendChat(ctx *fiber.Ctx) error
	DeleteSession(ctx *fiber.Ctx) error
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
	h.Get("/sessions", c.GetAllSession)
	h.Get("/chat-history", c.GetChatHistory)
	h.Post("/send-chat", c.SendChat)
	h.Delete("/delete-session", c.DeleteSession)
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

func (c *chatbotController) GetChatHistory(ctx *fiber.Ctx) error {

	idStr := ctx.Query("chat_session_id", "")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	res, err := c.chatbotService.GetChatHistory(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success get chat history", res))
}

func (c *chatbotController) SendChat(ctx *fiber.Ctx) error {

	var req dto.SendChatRequest

	err := ctx.BodyParser(&req)
	if err != nil {
		return err
	}

	err = serverutils.ValidateRequest(req)
	if err != nil {
		return err
	}

	res, err := c.chatbotService.SendChat(ctx.Context(), &req)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success create chat", res))
}

func (c *chatbotController) DeleteSession(ctx *fiber.Ctx) error {

	var req dto.DeleteSessionRequest

	err := ctx.BodyParser(&req)
	if err != nil {
		return err
	}

	err = serverutils.ValidateRequest(req)
	if err != nil {
		return err
	}

	err = c.chatbotService.DeleteSession(ctx.Context(), &req)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse[any]("Success Delete Session Chat", nil))
}
