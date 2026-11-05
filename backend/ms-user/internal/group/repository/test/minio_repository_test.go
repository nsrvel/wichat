package repository_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
)

type stubMinio struct {
	lastKey         string
	lastContentType string
	publicBase      string
}

func (s *stubMinio) Get(context.Context, string) (minio.Object, error) {
	return minio.Object{}, minio.ErrDisabled
}

func (s *stubMinio) Upload(_ context.Context, key, contentType string, _ io.Reader, _ int64) error {
	s.lastKey = key
	s.lastContentType = contentType

	return nil
}

func (s *stubMinio) UploadMany(context.Context, []minio.Upload) error { return minio.ErrDisabled }
func (s *stubMinio) Copy(context.Context, string, string) error       { return minio.ErrDisabled }
func (s *stubMinio) Delete(context.Context, string) error             { return minio.ErrDisabled }

func (s *stubMinio) DeletePrefix(_ context.Context, prefix string) error {
	s.lastKey = prefix

	return nil
}

func (s *stubMinio) PublicURL(key string) string {
	key = strings.TrimPrefix(key, "/")
	if s.publicBase == "" {
		return "/" + key
	}

	return s.publicBase + "/" + key
}

func TestPutGroupIconKeyAndURL(t *testing.T) {
	stub := &stubMinio{publicBase: "https://cdn.example.com/bucket"}
	repo := grouprepo.NewMinioRepository(stub)

	url, err := repo.PutGroupIcon(context.Background(), "abc123", "image/png", bytes.NewReader(nil), 0)
	if err != nil {
		t.Fatal(err)
	}
	if stub.lastKey != "groups/abc123/icon.png" {
		t.Fatalf("key %q", stub.lastKey)
	}
	if stub.lastContentType != "image/png" {
		t.Fatalf("content-type %q", stub.lastContentType)
	}
	want := "https://cdn.example.com/bucket/groups/abc123/icon.png"
	if url != want {
		t.Fatalf("url %q want %q", url, want)
	}
}

func TestRemoveGroupIconPrefix(t *testing.T) {
	stub := &stubMinio{}
	repo := grouprepo.NewMinioRepository(stub)

	if err := repo.RemoveGroupIcon(context.Background(), "abc123"); err != nil {
		t.Fatal(err)
	}
	if stub.lastKey != "groups/abc123/icon." {
		t.Fatalf("prefix %q", stub.lastKey)
	}
}
