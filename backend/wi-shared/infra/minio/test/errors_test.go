package minio_test

import (
	"context"
	"errors"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
)

func TestErrDisabledSentinel(t *testing.T) {
	c, err := minio.New(minio.Config{}, testLog())
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Upload(context.Background(), "k", "text/plain", nil, 0); !errors.Is(err, minio.ErrDisabled) {
		t.Fatalf("upload: %v", err)
	}
}
