package repository

import (
	"context"
	"io"
	"path"
	"strings"

	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
)

// MinioRepository persists channel message attachments in object storage.
type MinioRepository interface {
	PutChannelAttachment(ctx context.Context, channelID, attachmentID, filename, contentType string, body io.Reader, size int64) (publicURL string, err error)
}

type minioRepository struct {
	client minio.Client
}

// NewMinioRepository uploads channel attachments via the shared MinIO client.
func NewMinioRepository(client minio.Client) MinioRepository {
	return minioRepository{client: client}
}

func (r minioRepository) PutChannelAttachment(
	ctx context.Context,
	channelID, attachmentID, filename, contentType string,
	body io.Reader,
	size int64,
) (string, error) {
	key := channelAttachmentObjectKey(channelID, attachmentID, filename)
	if err := r.client.Upload(ctx, key, contentType, body, size); err != nil {
		return "", err
	}

	return r.client.PublicURL(key), nil
}

func channelAttachmentObjectKey(channelID, attachmentID, filename string) string {
	cleanName := path.Base(strings.ReplaceAll(strings.TrimSpace(filename), "\\", "/"))
	if cleanName == "" || cleanName == "." || cleanName == ".." || cleanName == "/" {
		cleanName = "file"
	}

	return "channels/" + channelID + "/attachments/" + attachmentID + "/" + cleanName
}
