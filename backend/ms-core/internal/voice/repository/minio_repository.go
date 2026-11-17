package repository

import "github.com/wichat/wichat/backend/wi-shared/infra/minio"

// MinioRepository resolves public/download URLs for stored call recordings.
type MinioRepository interface {
	RecordingURL(fileKey string) string
}

type minioRepository struct {
	client minio.Client
}

// NewMinioRepository wraps the shared MinIO client for call recordings.
func NewMinioRepository(client minio.Client) MinioRepository {
	return minioRepository{client: client}
}

func (r minioRepository) RecordingURL(fileKey string) string {
	if r.client == nil || fileKey == "" {
		return ""
	}
	return r.client.PublicURL(fileKey)
}
