package service

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/repository"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type IExampleService interface {
	HelloWorld(ctx context.Context, req *dto.HelloWorldRequest) (*dto.HelloWorldResponse, error)
	UploadFile(ctx context.Context, fileName string, fileContent io.ReadSeeker) (string, error)
	GetFileUrl(ctx context.Context, fileName string) (string, error)
}

type exampleService struct {
	exampleRepository repository.IExampleRepository
	s3Client          *s3.Client
}

func NewExampleService(exampleRepository repository.IExampleRepository, s3Client *s3.Client) IExampleService {
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
	bucketName := "test-folder"

	// 1. Deteksi Content-Type secara akurat (berdasarkan isi file)
	// Kita baca 512 byte pertama untuk menentukan tipe file yang asli
	buffer := make([]byte, 512)
	_, err := fileContent.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	contentType := http.DetectContentType(buffer)

	// Kembalikan pointer pembaca ke awal file setelah dibaca buffer-nya
	fileContent.Seek(0, io.SeekStart)

	// 2. Sanitasi & Unifikasi Nama File
	// Ganti spasi dengan underscore dan tambahkan timestamp agar tidak menimpa file lama
	ext := filepath.Ext(fileName)
	nameOnly := strings.TrimSuffix(fileName, ext)
	cleanName := strings.ReplaceAll(nameOnly, " ", "_")
	safeFileName := fmt.Sprintf("%d_%s%s", time.Now().Unix(), cleanName, ext)

	// 3. Upload dengan Metadata ContentType
	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(safeFileName),
		Body:        fileContent,
		ContentType: aws.String(contentType), // KUNCI: Agar muncul di browser
	})

	if err != nil {
		return "", err
	}

	return safeFileName, nil
}

func (s *exampleService) GetFileUrl(ctx context.Context, fileName string) (string, error) {
	bucketName := "test-folder"

	// 1. Inisialisasi Presign Client
	presignClient := s3.NewPresignClient(s.s3Client)

	// 2. Tentukan waktu kadaluarsa (misal: 15 menit)
	expiration := 15 * time.Minute

	// 3. Minta URL untuk operasi GetObject
	// Gunakan functional options s3.WithPresignExpires untuk mengatur durasi
	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
	}, s3.WithPresignExpires(expiration))

	if err != nil {
		return "", err
	}

	// Hasilnya ada di field URL
	return presignedReq.URL, nil
}
