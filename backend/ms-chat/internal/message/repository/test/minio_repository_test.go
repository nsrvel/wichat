package repository_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/ms-chat/internal/message/repository"
	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
)

type stubMinioClient struct {
	uploadKey         string
	uploadContentType string
	uploadSize        int64
	uploadErr         error
}

func (s *stubMinioClient) Get(context.Context, string) (minio.Object, error) {
	return minio.Object{}, nil
}

func (s *stubMinioClient) Upload(_ context.Context, key, contentType string, _ io.Reader, size int64) error {
	s.uploadKey = key
	s.uploadContentType = contentType
	s.uploadSize = size
	return s.uploadErr
}

func (s *stubMinioClient) UploadMany(context.Context, []minio.Upload) error {
	return nil
}

func (s *stubMinioClient) Copy(context.Context, string, string) error {
	return nil
}

func (s *stubMinioClient) Delete(context.Context, string) error {
	return nil
}

func (s *stubMinioClient) DeletePrefix(context.Context, string) error {
	return nil
}

func (s *stubMinioClient) PublicURL(key string) string {
	return "https://cdn.example.test/" + key
}

func TestMinioRepositoryPutChannelAttachment(t *testing.T) {
	t.Run("uploads under channel attachment key and returns public URL", func(t *testing.T) {
		stub := &stubMinioClient{}
		repo := repository.NewMinioRepository(stub)

		url, err := repo.PutChannelAttachment(
			context.Background(),
			"ch-1",
			"att-1",
			"report.pdf",
			"application/pdf",
			strings.NewReader("pdf-bytes"),
			9,
		)
		if err != nil {
			t.Fatalf("PutChannelAttachment() error = %v", err)
		}

		wantKey := "channels/ch-1/attachments/att-1/report.pdf"
		if stub.uploadKey != wantKey {
			t.Fatalf("uploadKey = %q, want %q", stub.uploadKey, wantKey)
		}
		if stub.uploadContentType != "application/pdf" {
			t.Fatalf("uploadContentType = %q, want application/pdf", stub.uploadContentType)
		}
		if stub.uploadSize != 9 {
			t.Fatalf("uploadSize = %d, want 9", stub.uploadSize)
		}
		if url != "https://cdn.example.test/"+wantKey {
			t.Fatalf("url = %q, want %q", url, "https://cdn.example.test/"+wantKey)
		}
	})

	t.Run("normalizes empty or traversal filename", func(t *testing.T) {
		stub := &stubMinioClient{}
		repo := repository.NewMinioRepository(stub)

		_, err := repo.PutChannelAttachment(
			context.Background(),
			"ch-1",
			"att-2",
			"../..",
			"text/plain",
			strings.NewReader("hi"),
			2,
		)
		if err != nil {
			t.Fatalf("PutChannelAttachment() error = %v", err)
		}
		wantKey := "channels/ch-1/attachments/att-2/file"
		if stub.uploadKey != wantKey {
			t.Fatalf("uploadKey = %q, want %q", stub.uploadKey, wantKey)
		}
	})

	t.Run("returns upload error", func(t *testing.T) {
		stub := &stubMinioClient{uploadErr: errors.New("upload failed")}
		repo := repository.NewMinioRepository(stub)

		_, err := repo.PutChannelAttachment(
			context.Background(),
			"ch-1",
			"att-3",
			"photo.png",
			"image/png",
			strings.NewReader("png"),
			3,
		)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
