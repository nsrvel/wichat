package pgErr

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

// MapNotFound maps sql.ErrNoRows to constants.ErrNotFound.
func MapNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return constants.ErrNotFound
	}

	return err
}

// MapUnique maps Postgres unique violation (23505) to constants.ErrConflict.
func MapUnique(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return constants.ErrConflict
	}

	return err
}
