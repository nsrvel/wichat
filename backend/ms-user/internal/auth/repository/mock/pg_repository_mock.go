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

// NewPgRepositoryMock is an in-memory PgRepository backed by shared state.
func NewPgRepositoryMock(st *usermock.State) repository.PgRepository {
	return &pgRepositoryMock{st: st}
}

func (m *pgRepositoryMock) InsertSession(_ context.Context, session models.Session, refresh models.RefreshToken) error {
	return m.st.InsertSessionRecord(session, refresh)
}

func (m *pgRepositoryMock) FindRefreshByHash(_ context.Context, hash []byte) (models.RefreshToken, error) {
	return m.st.FindRefreshByHash(hash)
}

func (m *pgRepositoryMock) FindSession(_ context.Context, id string) (models.Session, error) {
	return m.st.FindSession(id)
}

func (m *pgRepositoryMock) MarkRefreshUsed(_ context.Context, currentHash, nextHash []byte, nextID, sessionID string, expiresAt, now time.Time) error {
	return m.st.MarkRefreshUsed(currentHash, nextHash, nextID, sessionID, expiresAt, now)
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
