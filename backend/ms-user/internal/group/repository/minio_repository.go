package repository

import (
	"context"
	"io"
	"strings"

	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
)

// MinioRepository persists group branding assets (e.g. icons) in object storage.
type MinioRepository interface {
	PutGroupIcon(ctx context.Context, groupPublicID, contentType string, body io.Reader, size int64) (publicURL string, err error)
	RemoveGroupIcon(ctx context.Context, groupPublicID string) error
	PutGroupBanner(ctx context.Context, groupPublicID, contentType string, body io.Reader, size int64) (publicURL string, err error)
	RemoveGroupBanner(ctx context.Context, groupPublicID string) error
}

type minioRepository struct {
	client minio.Client
}

// NewMinioRepository uploads group assets via the shared MinIO client.
func NewMinioRepository(client minio.Client) MinioRepository {
	return minioRepository{client: client}
}

func (r minioRepository) PutGroupIcon(ctx context.Context, groupPublicID, contentType string, body io.Reader, size int64) (string, error) {
	ext := iconExtension(contentType)
	key := groupIconObjectKey(groupPublicID, ext)
	if err := r.client.Upload(ctx, key, contentType, body, size); err != nil {
		return "", err
	}

	return r.client.PublicURL(key), nil
}

func (r minioRepository) RemoveGroupIcon(ctx context.Context, groupPublicID string) error {
	return r.client.DeletePrefix(ctx, "groups/"+groupPublicID+"/icon.")
}

func (r minioRepository) PutGroupBanner(ctx context.Context, groupPublicID, contentType string, body io.Reader, size int64) (string, error) {
	ext := iconExtension(contentType)
	key := groupBannerObjectKey(groupPublicID, ext)
	if err := r.client.Upload(ctx, key, contentType, body, size); err != nil {
		return "", err
	}

	return r.client.PublicURL(key), nil
}

func (r minioRepository) RemoveGroupBanner(ctx context.Context, groupPublicID string) error {
	return r.client.DeletePrefix(ctx, "groups/"+groupPublicID+"/banner.")
}

func groupBannerObjectKey(groupPublicID, ext string) string {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	if ext == "" {
		ext = "bin"
	}

	return "groups/" + groupPublicID + "/banner." + ext
}

func groupIconObjectKey(groupPublicID, ext string) string {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	if ext == "" {
		ext = "bin"
	}

	return "groups/" + groupPublicID + "/icon." + ext
}

func iconExtension(contentType string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg", "image/jpg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	case "image/gif":
		return "gif"
	default:
		return "bin"
	}
}
