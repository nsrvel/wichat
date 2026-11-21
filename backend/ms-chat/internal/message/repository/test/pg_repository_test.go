package repository_test

import (
	"testing"

	"github.com/wichat/wichat/backend/ms-chat/internal/message/repository"
)

func TestNewPgRepository(t *testing.T) {
	repo := repository.NewPgRepository(nil, nil, nil)
	if repo == nil {
		t.Fatal("expected non-nil PgRepository")
	}
}
