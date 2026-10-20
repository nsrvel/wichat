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
	mu        sync.Mutex
	Users     []models.User
	Sessions  []models.Session
	Refreshes []models.RefreshToken
	Links     []googleLink
	Settings  models.InstanceSettings
	Groups    []models.Group
	Members   []models.GroupMember
}

// NewState returns empty mock storage.
func NewState() *State {
	return &State{}
}

// SetInstanceSettings preloads instance policy (e.g. guest enabled).
func (s *State) SetInstanceSettings(settings models.InstanceSettings) {
	s.mu.Lock()
	s.Settings = settings
	s.mu.Unlock()
}

func defaultInstanceSettings() models.InstanceSettings {
	return models.InstanceSettings{
		UsernameEnabled:  true,
		RegistrationMode: constants.RegistrationInviteOnly,
		GuestEnabled:     false,
		SetupCompleted:   false,
	}
}

func (s *State) InsertSessionRecord(session models.Session, refresh models.RefreshToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	refresh.TokenHash = cloneBytes(refresh.TokenHash)
	s.Sessions = append(s.Sessions, session)
	s.Refreshes = append(s.Refreshes, refresh)

	return nil
}

func (s *State) FindRefreshByHash(hash []byte) (models.RefreshToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, rec := range s.Refreshes {
		if bytes.Equal(rec.TokenHash, hash) {
			rec.TokenHash = cloneBytes(rec.TokenHash)

			return rec, nil
		}
	}

	return models.RefreshToken{}, constants.ErrNotFound
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

func (s *State) MarkRefreshUsed(currentHash, nextHash []byte, nextID, sessionID string, expiresAt, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.Refreshes {
		if !bytes.Equal(s.Refreshes[i].TokenHash, currentHash) {
			continue
		}
		if s.Refreshes[i].UsedAt != nil {
			return constants.ErrAlreadyUsed
		}
		used := now
		s.Refreshes[i].UsedAt = &used
		s.Refreshes = append(s.Refreshes, models.RefreshToken{
			ID:        nextID,
			SessionID: sessionID,
			TokenHash: cloneBytes(nextHash),
			CreatedAt: now,
		})
		for j := range s.Sessions {
			if s.Sessions[j].ID == sessionID {
				s.Sessions[j].ExpiresAt = expiresAt
				s.Sessions[j].LastUsedAt = now
			}
		}

		return nil
	}

	return constants.ErrNotFound
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

	keptRefresh := s.Refreshes[:0]
	for _, rec := range s.Refreshes {
		if _, drop := expired[rec.SessionID]; drop {
			continue
		}
		keptRefresh = append(keptRefresh, rec)
	}
	s.Refreshes = keptRefresh

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

func (m *pgRepositoryMock) GetFirstGroup(context.Context) (models.Group, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if len(m.st.Groups) == 0 {
		return models.Group{}, constants.ErrNotFound
	}

	return m.st.Groups[0], nil
}

func (m *pgRepositoryMock) CompleteInstanceSetup(_ context.Context, ownerUserID string, payload models.InstanceSetupPayload, now time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if m.st.Settings.SetupCompleted || len(m.st.Groups) > 0 {
		return constants.ErrNotFound
	}

	m.st.Settings = models.InstanceSettings{
		UsernameEnabled:  payload.UsernameEnabled,
		RegistrationMode: payload.RegistrationMode,
		GuestEnabled:     payload.GuestEnabled,
		SetupCompleted:   true,
	}
	m.st.Groups = append(m.st.Groups, models.Group{
		ID:        payload.GroupID,
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
			for _, existing := range m.st.Users {
				if existing.ID == link.userID {
					return existing, nil
				}
			}

			return models.User{}, constants.ErrNotFound
		}
	}
	for _, existing := range m.st.Users {
		if existing.Email == user.Email {
			m.st.Links = append(m.st.Links, googleLink{provider: provider, subject: subject, userID: existing.ID})

			return existing, nil
		}
	}

	m.st.Users = append(m.st.Users, user)
	m.st.Links = append(m.st.Links, googleLink{provider: provider, subject: subject, userID: user.ID})

	return user, nil
}

func (m *pgRepositoryMock) ListGroupsForUser(_ context.Context, userID string) ([]models.Group, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	return groupsForUser(m.st, userID), nil
}

func (m *pgRepositoryMock) ListJoinableGroups(_ context.Context, userID string) ([]models.Group, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	memberOf := map[string]bool{}
	for _, mem := range m.st.Members {
		if mem.UserID == userID {
			memberOf[mem.GroupID] = true
		}
	}
	var out []models.Group
	for _, g := range m.st.Groups {
		if !memberOf[g.ID] {
			out = append(out, g)
		}
	}

	return out, nil
}

func (m *pgRepositoryMock) FindGroupByID(_ context.Context, id string) (models.Group, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for _, g := range m.st.Groups {
		if g.ID == id {
			return g, nil
		}
	}

	return models.Group{}, constants.ErrNotFound
}

func (m *pgRepositoryMock) FindGroupMember(_ context.Context, groupID, userID string) (models.GroupMember, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for _, mem := range m.st.Members {
		if mem.GroupID == groupID && mem.UserID == userID {
			return mem, nil
		}
	}

	return models.GroupMember{}, constants.ErrNotFound
}

func (m *pgRepositoryMock) CreateGroupWithOwner(_ context.Context, group models.Group, owner models.GroupMember) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	m.st.Groups = append(m.st.Groups, group)
	m.st.Members = append(m.st.Members, owner)

	return nil
}

func (m *pgRepositoryMock) InsertGroupMember(_ context.Context, member models.GroupMember) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for _, mem := range m.st.Members {
		if mem.GroupID == member.GroupID && mem.UserID == member.UserID {
			return constants.ErrConflict
		}
	}
	m.st.Members = append(m.st.Members, member)

	return nil
}

func groupsForUser(st *State, userID string) []models.Group {
	memberOf := map[string]bool{}
	for _, mem := range st.Members {
		if mem.UserID == userID {
			memberOf[mem.GroupID] = true
		}
	}
	var out []models.Group
	for _, g := range st.Groups {
		if memberOf[g.ID] {
			out = append(out, g)
		}
	}

	return out
}

func cloneBytes(in []byte) []byte {
	out := make([]byte, len(in))
	copy(out, in)

	return out
}
