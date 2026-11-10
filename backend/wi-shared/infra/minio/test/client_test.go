package minio_test

import (
	"context"
	"errors"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
)

func testLog() logger.Logger {
	return logger.New(logger.Options{Service: "minio-test", Env: "test", Level: "error"})
}

func TestNewDisabledWhenEndpointEmpty(t *testing.T) {
	c, err := minio.New(minio.Config{}, testLog())
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Upload(context.Background(), "k", "text/plain", nil, 0); !errors.Is(err, minio.ErrDisabled) {
		t.Fatalf("upload: %v", err)
	}
	if _, err := c.Get(context.Background(), "k"); !errors.Is(err, minio.ErrDisabled) {
		t.Fatalf("get: %v", err)
	}
}

func TestDisabledPublicURL(t *testing.T) {
	c, err := minio.New(minio.Config{}, testLog())
	if err != nil {
		t.Fatal(err)
	}
	if c.PublicURL("groups/x/icon.png") != "/groups/x/icon.png" {
		t.Fatalf("got %q", c.PublicURL("groups/x/icon.png"))
	}
}

func TestDisabledClientBulkOps(t *testing.T) {
	c, err := minio.New(minio.Config{}, testLog())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	if err := c.UploadMany(ctx, []minio.Upload{{Key: "a", ContentType: "text/plain", Body: nil, Size: 0}}); !errors.Is(err, minio.ErrDisabled) {
		t.Fatalf("upload many: %v", err)
	}
	if err := c.Copy(ctx, "a", "b"); !errors.Is(err, minio.ErrDisabled) {
		t.Fatalf("copy: %v", err)
	}
	if err := c.Delete(ctx, "a"); !errors.Is(err, minio.ErrDisabled) {
		t.Fatalf("delete: %v", err)
	}
	if err := c.DeletePrefix(ctx, "p/"); !errors.Is(err, minio.ErrDisabled) {
		t.Fatalf("delete prefix: %v", err)
	}
}

func TestPublicURLWithBase(t *testing.T) {
	c, err := minio.New(minio.Config{
		Endpoint:      "localhost:9000",
		Bucket:        "assets",
		PublicBaseURL: "https://cdn.example.com/assets",
	}, testLog())
	if err != nil {
		t.Fatal(err)
	}
	got := c.PublicURL("/groups/abc/icon.png")
	want := "https://cdn.example.com/assets/groups/abc/icon.png"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
