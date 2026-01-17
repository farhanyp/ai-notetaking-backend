package controller

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type INoteController interface {
	RegisterRoutes(r fiber.Router)
	SemanticSearch(ctx *fiber.Ctx) error
	Create(ctx *fiber.Ctx) error
	Show(ctx *fiber.Ctx) error
	Update(ctx *fiber.Ctx) error
	Delete(ctx *fiber.Ctx) error
	GetExtractPreview(ctx *fiber.Ctx) error
	ConfirmExtraction(ctx *fiber.Ctx) error
}

type noteController struct {
	service service.INoteService
}

func NewNoteController(service service.INoteService) INoteController {
	return &noteController{service: service}
}

func (c *noteController) RegisterRoutes(r fiber.Router) {
	h := r.Group("/v1")
	h.Post("/note/create", c.Create)
	h.Get("/semantic-search", c.SemanticSearch)
	h.Get("/note/:id", c.Show)
	h.Put("/note/:id", c.Update)
	h.Delete("/note/:id", c.Delete)
	h.Put("/note/:id/move", c.Move)
	h.Get("/note/:id/extract-preview", c.GetExtractPreview)
	h.Put("/note/:id/confirm-extraction", c.ConfirmExtraction)

}

func (c *noteController) Create(ctx *fiber.Ctx) error {
	var req dto.CreateNoteRequest
	if err := ctx.BodyParser(&req); err != nil {
		return err
	}

	err := serverutils.ValidateRequest(req)
	if err != nil {
		return err
	}

	res, err := c.service.Create(ctx.Context(), &req)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success Creae Note", res))
}

func (c *noteController) Show(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, _ := uuid.Parse(idParam)

	res, err := c.service.Show(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success", res))
}

func (c *noteController) SemanticSearch(ctx *fiber.Ctx) error {
	q := ctx.Query("q", "")

	res, err := c.service.SemanticSearch(ctx.Context(), q)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success", res))
}

func (c *noteController) Update(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, _ := uuid.Parse(idParam)

	var req dto.UpdateNoteRequest
	if err := ctx.BodyParser(&req); err != nil {
		return err
	}

	req.Id = id

	err := serverutils.ValidateRequest(req)
	if err != nil {
		return err
	}

	res, err := c.service.Update(ctx.Context(), &req)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success Updated Note", res))
}

func (c *noteController) Delete(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, _ := uuid.Parse(idParam)

	err := c.service.Delete(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse[any]("Success Delete Note", nil))
}

func (c *noteController) Move(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, _ := uuid.Parse(idParam)

	var req dto.MoveNoteRequest
	if err := ctx.BodyParser(&req); err != nil {
		return err
	}
	req.Id = id

	err := serverutils.ValidateRequest(req)
	if err != nil {
		return err
	}

	res, err := c.service.Move(ctx.Context(), &req)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success Move Notebook", res))
}

func (c *noteController) GetExtractPreview(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid UUID format")
	}

	// Memanggil service untuk download dan ekstraksi AI
	extractedText, err := c.service.ExtractPreview(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success extract preview", fiber.Map{
		"note_id":        id,
		"extracted_text": extractedText,
	}))
}

func (c *noteController) ConfirmExtraction(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid UUID format")
	}

	var req dto.UpdateNoteRequest
	if err := ctx.BodyParser(&req); err != nil {
		return err
	}
	req.Id = id

	// Validasi konten teks hasil edit user
	err = serverutils.ValidateRequest(req)
	if err != nil {
		return err
	}

	// Update DB dan trigger re-embedding
	res, err := c.service.UpdateFromExtraction(ctx.Context(), &req)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success confirm and update extraction", res))
}
