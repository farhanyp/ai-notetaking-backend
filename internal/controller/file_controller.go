package controller

import (
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type IFileController interface {
	RegisterRoutes(r fiber.Router)
	UploadToGarage(ctx *fiber.Ctx) error
	GetFileURL(ctx *fiber.Ctx) error
}

type fileController struct {
	service service.IFileService
}

func NewFileController(service service.IFileService) IFileController {
	return &fileController{service: service}
}

func (c *fileController) RegisterRoutes(r fiber.Router) {
	h := r.Group("/v1")
	h.Post("/upload", c.UploadToGarage)
	h.Get("/get-file", c.GetFileURL)
}

func (c *fileController) UploadToGarage(ctx *fiber.Ctx) error {
	noteIdStr := ctx.FormValue("note_id")
	noteId, err := uuid.Parse(noteIdStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Note ID tidak valid")
	}

	fileHeader, err := ctx.FormFile("document")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "File document diperlukan")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	res, err := c.service.UploadFile(ctx.Context(), noteId, fileHeader.Filename, file)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("File berhasil diunggah dan disimpan", res))
}

func (c *fileController) GetFileURL(ctx *fiber.Ctx) error {
	fileName := ctx.Query("name")

	// Panggil Service
	url, err := c.service.GetFileUrl(ctx.Context(), fileName)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Berhasil mendapatkan URL file", url))
}
