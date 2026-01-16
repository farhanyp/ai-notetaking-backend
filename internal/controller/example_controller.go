package controller

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/internal/service"

	"github.com/gofiber/fiber/v2"
)

type IExampleController interface {
	RegisterRoutes(r fiber.Router)
	HelloWorld(ctx *fiber.Ctx) error
	UploadToGarage(ctx *fiber.Ctx) error
	GetFileURL(ctx *fiber.Ctx) error
}

type exampleController struct {
	service service.IExampleService
}

func NewExampleController(service service.IExampleService) IExampleController {
	return &exampleController{service: service}
}

func (c *exampleController) RegisterRoutes(r fiber.Router) {
	h := r.Group("/v1")
	h.Post("/hello-world", c.HelloWorld)
	h.Post("/upload", c.UploadToGarage)
	h.Get("/get-file", c.GetFileURL)
}

func (c *exampleController) HelloWorld(ctx *fiber.Ctx) error {
	var req dto.HelloWorldRequest
	if err := ctx.BodyParser(&req); err != nil {
		return err
	}

	err := serverutils.ValidateRequest(req)
	if err != nil {
		return err
	}

	res, err := c.service.HelloWorld(ctx.Context(), &req)
	if err != nil {
		return err
	}

	return ctx.JSON(serverutils.SuccessResponse("Success", res))
}

func (c *exampleController) UploadToGarage(ctx *fiber.Ctx) error {
	// 1. Ambil file dari form-data dengan key "document"
	fileHeader, err := ctx.FormFile("document")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Gagal mengambil file, pastikan key adalah 'document'",
		})
	}

	// 2. Buka file untuk mendapatkan stream datanya
	file, err := fileHeader.Open()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal membuka file",
		})
	}
	defer file.Close()

	// 3. Panggil Service untuk proses upload
	// Kita mengirim: context, nama file, dan isi file
	result, err := c.service.UploadFile(ctx.Context(), fileHeader.Filename, file)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// 4. Return response sukses
	return ctx.JSON(serverutils.SuccessResponse("File berhasil diunggah ke Garage S3", result))
}

func (c *exampleController) GetFileURL(ctx *fiber.Ctx) error {
	// Ambil nama file dari query parameter, misal: /get-url?name=foto.jpg
	fileName := ctx.Query("name")

	if fileName == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query parameter 'name' wajib diisi",
		})
	}

	// Panggil Service
	url, err := c.service.GetFileUrl(ctx.Context(), fileName)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.JSON(serverutils.SuccessResponse("Berhasil mendapatkan URL file", url))
}
