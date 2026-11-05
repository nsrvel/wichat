package mock

import (
	"context"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
)

type pgRepositoryMock struct {
	st *usermock.State
}

// NewPgRepositoryMock is an in-memory SessionRepository backed by shared state.
func NewPgRepositoryMock(st *usermock.State) repository.SessionRepository {
	return &pgRepositoryMock{st: st}
}

func (m *pgRepositoryMock) InsertSession(_ context.Context, session models.Session) error {
	return m.st.InsertSessionRecord(session)
}

func (m *pgRepositoryMock) FindSession(_ context.Context, id string) (models.Session, error) {
	return m.st.FindSession(id)
}

func (m *pgRepositoryMock) FindSessionByRefreshHash(_ context.Context, hash []byte, now time.Time) (models.Session, error) {
	return m.st.FindSessionByRefreshHash(hash, now)
}

func (m *pgRepositoryMock) TouchSession(_ context.Context, sessionID, refreshID string, refreshTokenHash []byte, expiresAt, lastUsedAt time.Time) error {
	return m.st.TouchSession(sessionID, refreshID, refreshTokenHash, expiresAt, lastUsedAt)
}

func (m *pgRepositoryMock) ListSessionIDsForUser(_ context.Context, userID string) ([]string, error) {
	return m.st.ListSessionIDsForUser(userID)
}

func (m *pgRepositoryMock) RevokeSession(_ context.Context, sessionID string, now time.Time) error {
	return m.st.RevokeSession(sessionID, now)
}

func (m *pgRepositoryMock) RevokeAllSessionsForUser(_ context.Context, userID string, now time.Time) error {
	return m.st.RevokeAllSessionsForUser(userID, now)
}

func (m *pgRepositoryMock) DeleteExpired(_ context.Context, now time.Time) error {
	return m.st.DeleteExpiredSessions(now)
}
