package repository_test

import (
	"testing"

	"github.com/wichat/wichat/backend/ms-core/internal/voice/repository"
)

func TestNewPgRepository(t *testing.T) {
	if repo := repository.NewPgRepository(nil, nil); repo == nil {
		t.Fatal("expected non-nil PgRepository")
	}
}
