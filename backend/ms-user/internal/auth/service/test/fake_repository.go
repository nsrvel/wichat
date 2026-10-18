package service_test

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
)

type googleLink struct {
	provider string
	subject  string
	userID   string
}

type fakeRepo struct {
	mu        sync.Mutex
	users     []models.User
	sessions  []models.Session
	refreshes []models.RefreshToken
	links     []googleLink
}

func (f *fakeRepo) CountUsers(context.Context) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return len(f.users), nil
}

func (f *fakeRepo) CreateFirstUser(_ context.Context, user models.User) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if len(f.users) > 0 {
		return repository.ErrConflict
	}
	f.users = append(f.users, user)

	return nil
}

func (f *fakeRepo) FindUserByEmail(_ context.Context, email string) (models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, user := range f.users {
		if user.Email == email {
			return user, nil
		}
	}

	return models.User{}, repository.ErrNotFound
}

func (f *fakeRepo) FindUserByID(_ context.Context, id string) (models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, user := range f.users {
		if user.ID == id {
			return user, nil
		}
	}

	return models.User{}, repository.ErrNotFound
}

func (f *fakeRepo) InsertSession(_ context.Context, session models.Session, refresh models.RefreshToken) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	refresh.TokenHash = cloneBytes(refresh.TokenHash)
	f.sessions = append(f.sessions, session)
	f.refreshes = append(f.refreshes, refresh)

	return nil
}

func (f *fakeRepo) FindRefreshByHash(_ context.Context, hash []byte) (models.RefreshToken, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, rec := range f.refreshes {
		if bytes.Equal(rec.TokenHash, hash) {
			rec.TokenHash = cloneBytes(rec.TokenHash)

			return rec, nil
		}
	}

	return models.RefreshToken{}, repository.ErrNotFound
}

func (f *fakeRepo) FindSession(_ context.Context, id string) (models.Session, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, session := range f.sessions {
		if session.ID == id {
			return session, nil
		}
	}

	return models.Session{}, repository.ErrNotFound
}

func (f *fakeRepo) MarkRefreshUsed(_ context.Context, currentHash, nextHash []byte, nextID, sessionID string, expiresAt, now time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	for i := range f.refreshes {
		if !bytes.Equal(f.refreshes[i].TokenHash, currentHash) {
			continue
		}
		if f.refreshes[i].UsedAt != nil {
			return repository.ErrAlreadyUsed
		}
		used := now
		f.refreshes[i].UsedAt = &used
		f.refreshes = append(f.refreshes, models.RefreshToken{
			ID:        nextID,
			SessionID: sessionID,
			TokenHash: cloneBytes(nextHash),
			CreatedAt: now,
		})
		for j := range f.sessions {
			if f.sessions[j].ID == sessionID {
				f.sessions[j].ExpiresAt = expiresAt
				f.sessions[j].LastUsedAt = now
			}
		}

		return nil
	}

	return repository.ErrNotFound
}

func (f *fakeRepo) RevokeSession(_ context.Context, sessionID string, now time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	for i := range f.sessions {
		if f.sessions[i].ID == sessionID && f.sessions[i].RevokedAt == nil {
			revoked := now
			f.sessions[i].RevokedAt = &revoked
		}
	}

	return nil
}

func (f *fakeRepo) SaveGoogleUser(_ context.Context, user models.User, _, provider, subject string) (models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, link := range f.links {
		if link.provider == provider && link.subject == subject {
			for _, existing := range f.users {
				if existing.ID == link.userID {
					return existing, nil
				}
			}

			return models.User{}, repository.ErrNotFound
		}
	}
	for _, existing := range f.users {
		if existing.Email == user.Email {
			f.links = append(f.links, googleLink{provider: provider, subject: subject, userID: existing.ID})

			return existing, nil
		}
	}

	f.users = append(f.users, user)
	f.links = append(f.links, googleLink{provider: provider, subject: subject, userID: user.ID})

	return user, nil
}

func (f *fakeRepo) DeleteExpired(_ context.Context, now time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	keptSessions := f.sessions[:0]
	expired := map[string]struct{}{}
	for _, session := range f.sessions {
		if !session.ExpiresAt.After(now) {
			expired[session.ID] = struct{}{}
			continue
		}
		keptSessions = append(keptSessions, session)
	}
	f.sessions = keptSessions

	keptRefresh := f.refreshes[:0]
	for _, rec := range f.refreshes {
		if _, drop := expired[rec.SessionID]; drop {
			continue
		}
		keptRefresh = append(keptRefresh, rec)
	}
	f.refreshes = keptRefresh

	return nil
}

func cloneBytes(in []byte) []byte {
	out := make([]byte, len(in))
	copy(out, in)

	return out
}
