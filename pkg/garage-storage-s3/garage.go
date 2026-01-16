package garagestorages3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Config menyimpan konfigurasi yang dibutuhkan untuk koneksi ke Garage S3
type Config struct {
	AccessKey string
	SecretKey string
	Endpoint  string
	Region    string
}

// NewGarageClient membuat instance S3 client yang sudah dikonfigurasi untuk Garage
func NewGarageClient(cfg Config) (*s3.Client, error) {
	// 1. Buat Provider Kredensial Statis
	staticResolver := credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")

	// 2. Load AWS Config standar
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(staticResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat konfigurasi sdk: %w", err)
	}

	// 3. Inisialisasi S3 Client dengan opsi khusus Garage
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		o.UsePathStyle = true // Wajib untuk Garage lokal
	})

	return client, nil
}
