package repository_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
)

type stubMinioClient struct {
	uploadKey         string
	uploadContentType string
	uploadSize        int64
	uploadErr         error
	deletePrefixArg   string
	deletePrefixErr   error
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

func (s *stubMinioClient) DeletePrefix(_ context.Context, prefix string) error {
	s.deletePrefixArg = prefix
	return s.deletePrefixErr
}

func (s *stubMinioClient) PublicURL(key string) string {
	return "https://cdn.example.test/" + key
}

func TestMinioRepositoryPutUserAvatar(t *testing.T) {
	t.Run("uploads with mapped extension and returns public URL", func(t *testing.T) {
		cases := []struct {
			contentType string
			wantExt     string
		}{
			{"image/jpeg", "jpg"},
			{"image/jpg", "jpg"},
			{"image/png", "png"},
			{"image/webp", "webp"},
			{"image/gif", "gif"},
			{"application/octet-stream", "bin"},
		}

		for _, tc := range cases {
			stub := &stubMinioClient{}
			repo := repository.NewMinioRepository(stub)
			url, err := repo.PutUserAvatar(context.Background(), "u1", tc.contentType, strings.NewReader("data"), 4)
			if err != nil {
				t.Fatalf("PutUserAvatar(%s) error = %v", tc.contentType, err)
			}
			wantKey := "users/u1/avatar." + tc.wantExt
			if stub.uploadKey != wantKey {
				t.Fatalf("uploadKey = %q, want %q", stub.uploadKey, wantKey)
			}
			if url != "https://cdn.example.test/"+wantKey {
				t.Fatalf("PublicURL = %q", url)
			}
		}
	})

	t.Run("returns upload error", func(t *testing.T) {
		stub := &stubMinioClient{uploadErr: errors.New("upload failed")}
		repo := repository.NewMinioRepository(stub)
		if _, err := repo.PutUserAvatar(context.Background(), "u1", "image/png", strings.NewReader("data"), 4); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMinioRepositoryRemoveUserAvatar(t *testing.T) {
	stub := &stubMinioClient{}
	repo := repository.NewMinioRepository(stub)
	if err := repo.RemoveUserAvatar(context.Background(), "u42"); err != nil {
		t.Fatalf("RemoveUserAvatar error = %v", err)
	}
	if stub.deletePrefixArg != "users/u42/avatar." {
		t.Fatalf("deletePrefix = %q, want %q", stub.deletePrefixArg, "users/u42/avatar.")
	}
}
