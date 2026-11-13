package mock

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

type googleLink struct {
	provider string
	subject  string
	userID   string
}

// State is shared in-memory storage for user and auth repository mocks.
type State struct {
	mu          sync.Mutex
	Users       []models.User
	Sessions    []models.Session
	Links       []googleLink
	Settings    models.InstanceSettings
	Groups      []models.Group
	Members     []models.GroupMember
	Invites     []models.GroupInviteRow
	Channels    []models.Channel
	Roles       []models.GroupRole
	MemberRoles []models.GroupMemberRoleRow
	Bans        []models.GroupBanRow
}

// NewState returns empty mock storage.
func NewState() *State {
	return &State{}
}

// Lock guards shared mock state for cross-package test helpers.
func (s *State) Lock() { s.mu.Lock() }

// Unlock releases the mock state lock.
func (s *State) Unlock() { s.mu.Unlock() }

// SetInstanceSettings preloads instance policy (e.g. guest enabled).
func (s *State) SetInstanceSettings(settings models.InstanceSettings) {
	s.mu.Lock()
	s.Settings = settings
	s.mu.Unlock()
}

func defaultInstanceSettings() models.InstanceSettings {
	return models.InstanceSettings{
		RegistrationMode: constants.RegistrationInviteOnly,
		SetupCompleted:   false,
	}
}

func (s *State) InsertSessionRecord(session models.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Sessions = append(s.Sessions, session)

	return nil
}

func (s *State) FindSession(id string) (models.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, session := range s.Sessions {
		if session.ID == id {
			return session, nil
		}
	}

	return models.Session{}, constants.ErrNotFound
}

func (s *State) FindSessionByRefreshHash(hash []byte, now time.Time) (models.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, session := range s.Sessions {
		if session.RevokedAt != nil || !session.ExpiresAt.After(now) {
			continue
		}
		if bytes.Equal(session.RefreshTokenHash, hash) {
			return session, nil
		}
	}

	return models.Session{}, constants.ErrNotFound
}

func (s *State) TouchSession(sessionID, refreshID string, refreshTokenHash []byte, expiresAt, lastUsedAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.Sessions {
		if s.Sessions[i].ID == sessionID {
			s.Sessions[i].RefreshID = refreshID
			s.Sessions[i].RefreshTokenHash = refreshTokenHash
			s.Sessions[i].ExpiresAt = expiresAt
			s.Sessions[i].LastUsedAt = lastUsedAt
		}
	}

	return nil
}

func (s *State) ListSessionIDsForUser(userID string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var ids []string
	for _, session := range s.Sessions {
		if session.UserID == userID && session.RevokedAt == nil {
			ids = append(ids, session.ID)
		}
	}

	return ids, nil
}

func (s *State) RevokeSession(sessionID string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.Sessions {
		if s.Sessions[i].ID == sessionID && s.Sessions[i].RevokedAt == nil {
			revoked := now
			s.Sessions[i].RevokedAt = &revoked
		}
	}

	return nil
}

func (s *State) RevokeAllSessionsForUser(userID string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.Sessions {
		if s.Sessions[i].UserID == userID && s.Sessions[i].RevokedAt == nil {
			revoked := now
			s.Sessions[i].RevokedAt = &revoked
		}
	}

	return nil
}

func (s *State) DeleteExpiredSessions(now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	keptSessions := s.Sessions[:0]
	expired := map[string]struct{}{}
	for _, session := range s.Sessions {
		if !session.ExpiresAt.After(now) {
			expired[session.ID] = struct{}{}
			continue
		}
		keptSessions = append(keptSessions, session)
	}
	s.Sessions = keptSessions

	return nil
}

type pgRepositoryMock struct {
	st *State
}

// NewPgRepositoryMock is an in-memory PgRepository backed by state.
func NewPgRepositoryMock(st *State) repository.PgRepository {
	return &pgRepositoryMock{st: st}
}

func (m *pgRepositoryMock) CountUsers(context.Context) (int, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	return len(m.st.Users), nil
}

func (m *pgRepositoryMock) GetInstanceSettings(context.Context) (models.InstanceSettings, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if m.st.Settings == (models.InstanceSettings{}) {
		return defaultInstanceSettings(), nil
	}

	return m.st.Settings, nil
}

func (m *pgRepositoryMock) CompleteInstanceSetup(_ context.Context, ownerUserID string, payload models.InstanceSetupPayload, now time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if m.st.Settings.SetupCompleted || len(m.st.Groups) > 0 {
		return constants.ErrNotFound
	}

	m.st.Settings = models.InstanceSettings{
		RegistrationMode: payload.RegistrationMode,
		SetupCompleted:   true,
	}
	m.st.Groups = append(m.st.Groups, models.Group{
		ID:        payload.GroupID,
		PublicID:  payload.GroupPublicID,
		Name:      payload.GroupName,
		LogoURL:   payload.GroupLogoURL,
		CreatedAt: now,
		UpdatedAt: now,
	})
	m.st.Members = append(m.st.Members, models.GroupMember{
		GroupID:   payload.GroupID,
		UserID:    ownerUserID,
		Role:      constants.GroupRoleOwner,
		CreatedAt: now,
	})

	return nil
}

func (m *pgRepositoryMock) CreateFirstUser(_ context.Context, user models.User) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if len(m.st.Users) > 0 {
		return constants.ErrConflict
	}
	m.st.Users = append(m.st.Users, user)

	return nil
}

func (m *pgRepositoryMock) CreateUser(_ context.Context, user models.User) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for _, existing := range m.st.Users {
		if user.Email != "" && existing.Email == user.Email {
			return constants.ErrConflict
		}
		if user.Username != "" && existing.Username == user.Username {
			return constants.ErrConflict
		}
	}
	m.st.Users = append(m.st.Users, user)

	return nil
}

func (m *pgRepositoryMock) FindUserByEmail(_ context.Context, email string) (models.User, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for _, user := range m.st.Users {
		if user.Email == email {
			return user, nil
		}
	}

	return models.User{}, constants.ErrNotFound
}

func (m *pgRepositoryMock) FindUserByUsername(_ context.Context, username string) (models.User, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for _, user := range m.st.Users {
		if user.Username == username {
			return user, nil
		}
	}

	return models.User{}, constants.ErrNotFound
}

func (m *pgRepositoryMock) FindUserByLogin(_ context.Context, login string) (models.User, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	needle := strings.ToLower(strings.TrimSpace(login))
	for _, user := range m.st.Users {
		if strings.ToLower(user.Email) == needle || strings.ToLower(user.Username) == needle {
			return user, nil
		}
	}

	return models.User{}, constants.ErrNotFound
}

func (m *pgRepositoryMock) FindUserByID(_ context.Context, id string) (models.User, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for _, user := range m.st.Users {
		if user.ID == id {
			return user, nil
		}
	}

	return models.User{}, constants.ErrNotFound
}

func (m *pgRepositoryMock) SearchUsers(_ context.Context, query string, limit int, excludeGroupID string) ([]models.UserSearchHit, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" || limit <= 0 {
		return nil, nil
	}

	memberIDs := make(map[string]struct{})
	for _, mem := range m.st.Members {
		if mem.GroupID == excludeGroupID {
			memberIDs[mem.UserID] = struct{}{}
		}
	}

	var out []models.UserSearchHit
	for _, user := range m.st.Users {
		if _, inGroup := memberIDs[user.ID]; inGroup {
			continue
		}
		if strings.Contains(strings.ToLower(user.Username), query) ||
			strings.Contains(strings.ToLower(user.DisplayName), query) {
			out = append(out, models.UserSearchHit{
				UserID:      user.ID,
				Username:    user.Username,
				DisplayName: user.DisplayName,
				AvatarURL:   user.AvatarURL,
			})
			if len(out) >= limit {
				break
			}
		}
	}

	return out, nil
}

func (m *pgRepositoryMock) UpdateUserPasswordHash(_ context.Context, userID, hash string, now time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i := range m.st.Users {
		if m.st.Users[i].ID == userID {
			m.st.Users[i].PasswordHash = hash
			m.st.Users[i].UpdatedAt = now

			return nil
		}
	}

	return constants.ErrNotFound
}

func (m *pgRepositoryMock) SaveGoogleUser(_ context.Context, user models.User, _, provider, subject string) (models.User, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for _, link := range m.st.Links {
		if link.provider == provider && link.subject == subject {
			for i := range m.st.Users {
				if m.st.Users[i].ID == link.userID {
					if m.st.Users[i].AvatarURL == "" && user.AvatarURL != "" {
						m.st.Users[i].AvatarURL = user.AvatarURL
					}

					return m.st.Users[i], nil
				}
			}

			return models.User{}, constants.ErrNotFound
		}
	}
	for i := range m.st.Users {
		if m.st.Users[i].Email == user.Email {
			if m.st.Users[i].AvatarURL == "" && user.AvatarURL != "" {
				m.st.Users[i].AvatarURL = user.AvatarURL
			}
			m.st.Links = append(m.st.Links, googleLink{provider: provider, subject: subject, userID: m.st.Users[i].ID})

			return m.st.Users[i], nil
		}
	}

	m.st.Users = append(m.st.Users, user)
	m.st.Links = append(m.st.Links, googleLink{provider: provider, subject: subject, userID: user.ID})

	return user, nil
}
