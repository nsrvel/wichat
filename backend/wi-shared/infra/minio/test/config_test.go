package minio_test

import (
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
)

func TestNewRequiresBucketWhenEndpointSet(t *testing.T) {
	_, err := minio.New(minio.Config{Endpoint: "localhost:9000"}, testLog())
	if err == nil {
		t.Fatal("expected error")
	}
}
