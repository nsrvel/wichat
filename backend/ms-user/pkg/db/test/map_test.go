package pgErr_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	pgErr "github.com/wichat/wichat/backend/ms-user/pkg/db"
)

func TestMapNotFound(t *testing.T) {
	t.Parallel()

	if err := pgErr.MapNotFound(nil); err != nil {
		t.Fatal("nil")
	}
	if !errors.Is(pgErr.MapNotFound(sql.ErrNoRows), constants.ErrNotFound) {
		t.Fatal("no rows")
	}
	if pgErr.MapNotFound(errors.New("other")).Error() != "other" {
		t.Fatal("passthrough")
	}
}

func TestMapUnique(t *testing.T) {
	t.Parallel()

	pg := &pgconn.PgError{Code: "23505"}
	if !errors.Is(pgErr.MapUnique(pg), constants.ErrConflict) {
		t.Fatal("unique")
	}
	if pgErr.MapUnique(errors.New("x")).Error() != "x" {
		t.Fatal("passthrough")
	}
}
