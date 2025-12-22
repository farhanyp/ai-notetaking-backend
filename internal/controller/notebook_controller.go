package controller

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type INotebookController interface {
	RegisterRoutes(r fiber.Router)
	GetAll(ctx *fiber.Ctx) error
	Create(ctx *fiber.Ctx) error
	Show(ctx *fiber.Ctx) error
	Update(ctx *fiber.Ctx) error
	Delete(ctx *fiber.Ctx) error
	Move(ctx *fiber.Ctx) error
}

type notebookeController struct {
	service service.INotebookService
}

func NewNotebookController(service service.INotebookService) INotebookController {
	return &notebookeController{service: service}
}

func (c *notebookeController) RegisterRoutes(r fiber.Router) {
	h := r.Group("/v1")
	h.Get("/notebook", c.GetAll)
	h.Post("/notebook/create", c.Create)
	h.Get("/notebook/:id", c.Show)
	h.Put("/notebook/:id", c.Update)
	h.Delete("/notebook/:id", c.Delete)
	h.Put("/notebook/:id/move", c.Move)
}

func (c *notebookeController) GetAll(ctx *fiber.Ctx) error {

	res, err := c.service.GetAll(ctx.Context())
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Get List Notebook Success", res))
}

func (c *notebookeController) Create(ctx *fiber.Ctx) error {
	var req dto.CreateNotebookRequest
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

	return ctx.JSON(serverutils.SuccessResponse("Success", res))
}

func (c *notebookeController) Show(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, _ := uuid.Parse(idParam)

	res, err := c.service.Show(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success", res))
}

func (c *notebookeController) Update(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, _ := uuid.Parse(idParam)

	var req dto.UpdateNotebookRequest
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

	return ctx.JSON(serverutils.SuccessResponse("Success Updated Notebook", res))
}

func (c *notebookeController) Move(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, _ := uuid.Parse(idParam)

	var req dto.MoveNotebookRequest
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

func (c *notebookeController) Delete(ctx *fiber.Ctx) error {
	idParam := ctx.Params("id")
	id, _ := uuid.Parse(idParam)

	err := c.service.Delete(ctx.Context(), id)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse[any]("Success Delete Notebook", nil))
}