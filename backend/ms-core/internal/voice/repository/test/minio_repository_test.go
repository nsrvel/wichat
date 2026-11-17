package repository_test

import (
	"context"
	"io"
	"testing"

	"github.com/wichat/wichat/backend/ms-core/internal/voice/repository"
	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
)

type stubMinio struct{}

func (stubMinio) Get(context.Context, string) (minio.Object, error) { return minio.Object{}, nil }
func (stubMinio) Upload(context.Context, string, string, io.Reader, int64) error {
	return nil
}
func (stubMinio) UploadMany(context.Context, []minio.Upload) error { return nil }
func (stubMinio) Copy(context.Context, string, string) error       { return nil }
func (stubMinio) Delete(context.Context, string) error             { return nil }
func (stubMinio) DeletePrefix(context.Context, string) error       { return nil }
func (stubMinio) PublicURL(key string) string                      { return "https://cdn.example.test/" + key }

func TestMinioRepositoryRecordingURL(t *testing.T) {
	repo := repository.NewMinioRepository(stubMinio{})
	if got := repo.RecordingURL("recordings/a.mp4"); got != "https://cdn.example.test/recordings/a.mp4" {
		t.Fatalf("RecordingURL = %q", got)
	}
	if got := repo.RecordingURL(""); got != "" {
		t.Fatalf("expected empty for empty key, got %q", got)
	}
}
