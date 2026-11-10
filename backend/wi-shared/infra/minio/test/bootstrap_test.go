package minio_test

import (
	"context"
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
)

func TestEnsurePublicReadBucketNoOpWhenDisabled(t *testing.T) {
	c, err := minio.New(minio.Config{}, testLog())
	if err != nil {
		t.Fatal(err)
	}
	if err := minio.EnsurePublicReadBucket(context.Background(), c); err != nil {
		t.Fatal(err)
	}
}
