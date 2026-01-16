package garagestorages3

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type GarageS3 struct {
	Client *s3.Client
}

type Config struct {
	AccessKey string
	SecretKey string
	Endpoint  string
	Region    string
}

// NewGarageClient inisialisasi client utama
func NewGarageClient(cfg Config) (*GarageS3, error) {
	staticResolver := credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")

	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(staticResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load sdk config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true
	})

	return &GarageS3{Client: client}, nil
}

// ---------------------------------------------------------
// CORE FUNCTIONS
// ---------------------------------------------------------

// Upload mengunggah file dan mengembalikan nama file yang sudah aman
func (g *GarageS3) Upload(ctx context.Context, bucket, fileName string, content io.ReadSeeker) (string, error) {
	// Detect Content Type
	contentType, err := g.DetectMimeType(content)
	if err != nil {
		return "", err
	}

	// Generate Safe Name
	safeName := g.GenerateSafeFileName(fileName)

	_, err = g.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(safeName),
		Body:        content,
		ContentType: aws.String(contentType),
	})

	return safeName, err
}

// GetPresignedURL membuat link akses sementara (default 15 menit)
func (g *GarageS3) GetPresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(g.Client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))

	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// FileExists mengecek apakah file ada di storage
func (g *GarageS3) FileExists(ctx context.Context, bucket, key string) (bool, error) {
	_, err := g.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Jika error 404, file tidak ada
		return false, nil
	}
	return true, nil
}

// Delete menghapus file dari storage
func (g *GarageS3) Delete(ctx context.Context, bucket, key string) error {
	_, err := g.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

// ---------------------------------------------------------
// HELPER FUNCTIONS (Independent)
// ---------------------------------------------------------

// GenerateSafeFileName merubah nama menjadi unik (timestamp_nama_file.ext)
func (g *GarageS3) GenerateSafeFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	nameOnly := strings.TrimSuffix(originalName, ext)

	// Hapus karakter aneh dan ganti spasi jadi underscore
	cleanName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, nameOnly)

	return fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), cleanName, ext)
}

// DetectMimeType mendeteksi tipe file dari stream
func (g *GarageS3) DetectMimeType(content io.ReadSeeker) (string, error) {
	buffer := make([]byte, 512)
	_, err := content.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Reset kursor ke awal agar file bisa dibaca penuh saat upload
	_, _ = content.Seek(0, io.SeekStart)

	return http.DetectContentType(buffer), nil
}

// ValidateAllowedMime memvalidasi apakah tipe file diizinkan (e.g., ["image/jpeg", "application/pdf"])
func (g *GarageS3) ValidateAllowedMime(content io.ReadSeeker, allowedTypes []string) (bool, string, error) {
	mimeType, err := g.DetectMimeType(content)
	if err != nil {
		return false, "", err
	}

	for _, t := range allowedTypes {
		if mimeType == t {
			return true, mimeType, nil
		}
	}
	return false, mimeType, nil
}
