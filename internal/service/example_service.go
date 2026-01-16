package service

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/repository"
	garagestorages3 "ai-notetaking-be/pkg/garage-storage-s3"
	"context"
	"fmt"
	"io"
	"time"
)

type IExampleService interface {
	HelloWorld(ctx context.Context, req *dto.HelloWorldRequest) (*dto.HelloWorldResponse, error)
	UploadFile(ctx context.Context, fileName string, fileContent io.ReadSeeker) (string, error)
	GetFileUrl(ctx context.Context, fileName string) (string, error)
}

type exampleService struct {
	exampleRepository repository.IExampleRepository
	s3Client          *garagestorages3.GarageS3
}

func NewExampleService(exampleRepository repository.IExampleRepository, s3Client *garagestorages3.GarageS3) IExampleService {
	return &exampleService{
		exampleRepository: exampleRepository,
		s3Client:          s3Client,
	}
}

func (c *exampleService) HelloWorld(ctx context.Context, req *dto.HelloWorldRequest) (*dto.HelloWorldResponse, error) {
	_, err := c.exampleRepository.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.HelloWorldResponse{
		Message: fmt.Sprintf(`Hello %s`, req.Name),
	}, nil
}

func (s *exampleService) UploadFile(ctx context.Context, fileName string, fileContent io.ReadSeeker) (string, error) {
	// 1. Validasi MIME type menggunakan helper library
	isOk, mime, err := s.s3Client.ValidateAllowedMime(fileContent, []string{"image/jpeg", "image/png", "application/pdf"})
	if err != nil {
		return "", fmt.Errorf("gagal memvalidasi file: %w", err)
	}
	if !isOk {
		return "", fmt.Errorf("tipe file %s tidak diizinkan", mime)
	}

	// 2. Gunakan fungsi Upload dari library (otomatis sanitasi nama & set Content-Type)
	// Kita simpan ke bucket "test-folder"
	safeName, err := s.s3Client.Upload(ctx, "test-folder", fileName, fileContent)
	if err != nil {
		return "", fmt.Errorf("gagal mengunggah ke storage: %w", err)
	}

	return safeName, nil
}

func (s *exampleService) GetFileUrl(ctx context.Context, fileName string) (string, error) {
	// Panggil fungsi GetPresignedURL yang sudah tersedia di library
	// Ini jauh lebih bersih daripada menginisialisasi presignClient secara manual di sini
	url, err := s.s3Client.GetPresignedURL(ctx, "test-folder", fileName, 15*time.Minute)
	if err != nil {
		return "", fmt.Errorf("gagal mendapatkan URL file: %w", err)
	}

	return url, nil
}
