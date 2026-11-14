package repository

import (
	"context"
	"io"
	"strings"

	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
)

// MinioRepository persists user profile assets (avatars) in object storage.
type MinioRepository interface {
	PutUserAvatar(ctx context.Context, userID, contentType string, body io.Reader, size int64) (publicURL string, err error)
	RemoveUserAvatar(ctx context.Context, userID string) error
}

type minioRepository struct {
	client minio.Client
}

// NewMinioRepository uploads user assets via the shared MinIO client.
func NewMinioRepository(client minio.Client) MinioRepository {
	return minioRepository{client: client}
}

func (r minioRepository) PutUserAvatar(ctx context.Context, userID, contentType string, body io.Reader, size int64) (string, error) {
	ext := avatarExtension(contentType)
	key := userAvatarObjectKey(userID, ext)
	if err := r.client.Upload(ctx, key, contentType, body, size); err != nil {
		return "", err
	}

	return r.client.PublicURL(key), nil
}

func (r minioRepository) RemoveUserAvatar(ctx context.Context, userID string) error {
	return r.client.DeletePrefix(ctx, "users/"+userID+"/avatar.")
}

func userAvatarObjectKey(userID, ext string) string {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	if ext == "" {
		ext = "bin"
	}

	return "users/" + userID + "/avatar." + ext
}

func avatarExtension(contentType string) string {
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
