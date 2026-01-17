package service

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/internal/repository"
	garagestorages3 "ai-notetaking-be/pkg/garage-storage-s3"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
)

type IFileService interface {
	UploadFile(ctx context.Context, noteId uuid.UUID, fileName string, content io.ReadSeeker) (*dto.UploadFileResponse, error)
	GetFileUrl(ctx context.Context, fileName string) (string, error)
}

type fileService struct {
	noteRepository repository.INoteRepository
	fileRepository repository.IFileRepository
	s3Client       *garagestorages3.GarageS3
}

func NewFileService(
	noteRepository repository.INoteRepository,
	fileRepository repository.IFileRepository,
	s3Client *garagestorages3.GarageS3,
) IFileService {
	return &fileService{
		noteRepository: noteRepository,
		fileRepository: fileRepository,
		s3Client:       s3Client,
	}
}

func (s *fileService) UploadFile(ctx context.Context, noteId uuid.UUID, fileName string, content io.ReadSeeker) (*dto.UploadFileResponse, error) {
	_, err := s.noteRepository.GetById(ctx, noteId)
	if err != nil {
		return nil, err
	}

	safeFileName, err := s.s3Client.Upload(ctx, os.Getenv("BUCKET"), fileName, content)
	if err != nil {
		return nil, fmt.Errorf("gagal upload ke storage: %w", err)
	}

	mimeType, _ := s.s3Client.DetectMimeType(content)

	fileEntity := &entity.File{
		Id:          uuid.New(),
		FileName:    safeFileName,
		Bucket:      os.Getenv("BUCKET"),
		ContentType: mimeType,
		NoteId:      noteId,
		CreatedAt:   time.Now(),
	}

	err = s.fileRepository.Create(ctx, fileEntity)
	if err != nil {
		// 1. Log error secara detail ke console
		fmt.Printf("[ERROR] Database Create File: %v\n", err)

		// 2. Rollback file di S3 agar tidak menumpuk sampah
		deleteErr := s.s3Client.Delete(ctx, os.Getenv("BUCKET"), safeFileName)
		if deleteErr != nil {
			fmt.Printf("[CRITICAL] Gagal menghapus file yatim di S3: %v\n", deleteErr)
		}

		// 3. Kembalikan error asli agar terlihat di Response API
		return nil, fmt.Errorf("gagal menyimpan metadata file ke database: %w", err)
	}

	return &dto.UploadFileResponse{
		FileId:   fileEntity.Id,
		FileName: fileEntity.FileName,
	}, nil
}

func (s *fileService) GetFileUrl(ctx context.Context, fileName string) (string, error) {
	// 1. Check if the file name is empty before proceeding
	if fileName == "" {
		return "", serverutils.ErrBadRequest
	}

	// 2. Call GetPresignedURL from the library
	// We use a fixed bucket "test-folder" as per your setup
	url, err := s.s3Client.GetPresignedURL(ctx, os.Getenv("BUCKET"), fileName, 15*time.Minute)

	if err != nil {
		// LOG the technical details (e.g., S3 connection timeout, signature error)
		log.Printf("[S3 Service] Failed to generate Presigned URL for %s: %v", fileName, err)

		// RETURN a safe error to the client
		return "", serverutils.ErrInternal
	}

	return url, nil
}
