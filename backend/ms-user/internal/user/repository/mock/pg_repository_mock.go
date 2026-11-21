package mock

import (
	"bytes"
	"context"
	"slices"
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

type recoveryCodeMock struct {
	hash []byte
	used bool
}

// State is shared in-memory storage for user and auth repository mocks.
type State struct {
	mu                          sync.Mutex
	Users                       []models.User
	Sessions                    []models.Session
	Links                       []googleLink
	Groups                      []models.Group
	Members                     []models.GroupMember
	Invites                     []models.GroupInviteRow
	Channels                    []models.Channel
	ChannelCategories           []models.ChannelCategory
	CategoryPermissionOverrides []models.ChannelCategoryPermissionOverride
	ChannelPermissionOverrides  []models.ChannelPermissionOverride
	Roles                       []models.GroupRole
	MemberRoles                 []models.GroupMemberRoleRow
	Bans                        []models.GroupBanRow
	MFATOTPs                    map[string]models.MFATOTP
	WebAuthnCreds               []models.WebAuthnCredential
	RecoveryCodes               map[string][]recoveryCodeMock
	Relationships               []models.UserRelationship
}

// NewState returns empty mock storage.
func NewState() *State {
	return &State{}
}

// Lock guards shared mock state for cross-package test helpers.
func (s *State) Lock() { s.mu.Lock() }

// Unlock releases the mock state lock.
func (s *State) Unlock() { s.mu.Unlock() }

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

func (s *State) ListActiveSessionsForUser(userID string, now time.Time) ([]models.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var out []models.Session
	for _, session := range s.Sessions {
		if session.UserID == userID && session.RevokedAt == nil && session.ExpiresAt.After(now) {
			out = append(out, session)
		}
	}
	slices.SortFunc(out, func(a, b models.Session) int {
		return b.LastUsedAt.Compare(a.LastUsedAt)
	})

	return out, nil
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

func (s *State) RevokeOtherSessionsForUser(userID, exceptSessionID string, now time.Time) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var ids []string
	for i := range s.Sessions {
		if s.Sessions[i].UserID == userID && s.Sessions[i].ID != exceptSessionID && s.Sessions[i].RevokedAt == nil {
			revoked := now
			s.Sessions[i].RevokedAt = &revoked
			ids = append(ids, s.Sessions[i].ID)
		}
	}

	return ids, nil
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

func (m *pgRepositoryMock) CompleteInstanceSetup(_ context.Context, ownerUserID string, payload models.InstanceSetupPayload, now time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if len(m.st.Groups) > 0 {
		return constants.ErrConflict
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
		if user.DeletedAt != nil || user.DisabledAt != nil {
			continue
		}
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

func (m *pgRepositoryMock) DisableAccount(_ context.Context, userID string, now time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i := range m.st.Users {
		if m.st.Users[i].ID == userID && m.st.Users[i].DeletedAt == nil {
			t := now
			m.st.Users[i].DisabledAt = &t
			m.st.Users[i].UpdatedAt = now

			return nil
		}
	}

	return constants.ErrNotFound
}

func (m *pgRepositoryMock) SoftDeleteAccount(_ context.Context, userID, anonymizedUsername string, now time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	found := false
	for i := range m.st.Users {
		if m.st.Users[i].ID == userID && m.st.Users[i].DeletedAt == nil {
			t := now
			m.st.Users[i].DeletedAt = &t
			m.st.Users[i].DisabledAt = nil
			m.st.Users[i].Email = ""
			m.st.Users[i].Username = anonymizedUsername
			m.st.Users[i].DisplayName = "Deleted User"
			m.st.Users[i].PasswordHash = ""
			m.st.Users[i].AvatarURL = ""
			m.st.Users[i].Pronouns = ""
			m.st.Users[i].Bio = ""
			m.st.Users[i].BannerColor = ""
			m.st.Users[i].UpdatedAt = now
			found = true
			break
		}
	}
	if !found {
		return constants.ErrNotFound
	}

	links := m.st.Links[:0]
	for _, l := range m.st.Links {
		if l.userID != userID {
			links = append(links, l)
		}
	}
	m.st.Links = links

	members := m.st.Members[:0]
	for _, mem := range m.st.Members {
		if mem.UserID != userID {
			members = append(members, mem)
		}
	}
	m.st.Members = members

	delete(m.st.MFATOTPs, userID)
	delete(m.st.RecoveryCodes, userID)
	creds := m.st.WebAuthnCreds[:0]
	for _, c := range m.st.WebAuthnCreds {
		if c.UserID != userID {
			creds = append(creds, c)
		}
	}
	m.st.WebAuthnCreds = creds

	return nil
}

func (m *pgRepositoryMock) UpdateUserAvatarURL(_ context.Context, userID, avatarURL string, now time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i := range m.st.Users {
		if m.st.Users[i].ID == userID {
			m.st.Users[i].AvatarURL = avatarURL
			m.st.Users[i].UpdatedAt = now
			return nil
		}
	}
	return constants.ErrNotFound
}

func (m *pgRepositoryMock) UpdateUserProfile(_ context.Context, userID string, patch models.UserProfilePatch, now time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i := range m.st.Users {
		if m.st.Users[i].ID == userID {
			m.st.Users[i].DisplayName = patch.DisplayName
			m.st.Users[i].Pronouns = patch.Pronouns
			m.st.Users[i].Bio = patch.Bio
			m.st.Users[i].BannerColor = patch.BannerColor
			if patch.ClearAvatar {
				m.st.Users[i].AvatarURL = ""
			}
			m.st.Users[i].UpdatedAt = now
			return nil
		}
	}
	return constants.ErrNotFound
}

func (m *pgRepositoryMock) UpdateUserStatusNote(_ context.Context, userID string, patch models.UserStatusNotePatch, now time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i := range m.st.Users {
		if m.st.Users[i].ID == userID {
			m.st.Users[i].StatusNoteText = patch.Text
			m.st.Users[i].StatusNoteEmoji = patch.Emoji
			m.st.Users[i].StatusNoteExpiresAt = patch.ExpiresAt
			m.st.Users[i].UpdatedAt = now
			return nil
		}
	}
	return constants.ErrNotFound
}

func (m *pgRepositoryMock) UpdateUserPreferences(_ context.Context, userID string, patch models.UserPreferencesPatch, now time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i := range m.st.Users {
		if m.st.Users[i].ID == userID {
			m.st.Users[i].Locale = patch.Locale
			m.st.Users[i].TimeFormat = patch.TimeFormat
			m.st.Users[i].UpdatedAt = now
			return nil
		}
	}
	return constants.ErrNotFound
}

func (m *pgRepositoryMock) UpdateUsername(_ context.Context, userID, username string, now time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for _, u := range m.st.Users {
		if u.ID != userID && u.Username == username {
			return constants.ErrConflict
		}
	}
	for i := range m.st.Users {
		if m.st.Users[i].ID == userID {
			m.st.Users[i].Username = username
			m.st.Users[i].UpdatedAt = now
			return nil
		}
	}
	return constants.ErrNotFound
}

func (m *pgRepositoryMock) UpdateEmail(_ context.Context, userID, email string, now time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for _, u := range m.st.Users {
		if u.ID != userID && strings.EqualFold(u.Email, email) {
			return constants.ErrConflict
		}
	}
	for i := range m.st.Users {
		if m.st.Users[i].ID == userID {
			m.st.Users[i].Email = email
			m.st.Users[i].UpdatedAt = now
			return nil
		}
	}
	return constants.ErrNotFound
}

func (m *pgRepositoryMock) GetMFATOTP(_ context.Context, userID string) (models.MFATOTP, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if m.st.MFATOTPs == nil {
		return models.MFATOTP{}, constants.ErrNotFound
	}
	t, ok := m.st.MFATOTPs[userID]
	if !ok {
		return models.MFATOTP{}, constants.ErrNotFound
	}
	return t, nil
}

func (m *pgRepositoryMock) UpsertMFATOTP(_ context.Context, totp models.MFATOTP) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if m.st.MFATOTPs == nil {
		m.st.MFATOTPs = make(map[string]models.MFATOTP)
	}
	m.st.MFATOTPs[totp.UserID] = totp
	return nil
}

func (m *pgRepositoryMock) UpdateMFATOTPLastUsedStep(_ context.Context, userID string, step int64, now time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if m.st.MFATOTPs == nil {
		return nil
	}
	if t, ok := m.st.MFATOTPs[userID]; ok {
		t.LastUsedStep = step
		t.UpdatedAt = now
		m.st.MFATOTPs[userID] = t
	}
	return nil
}

func (m *pgRepositoryMock) DeleteMFATOTP(_ context.Context, userID string) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if _, ok := m.st.MFATOTPs[userID]; !ok {
		return constants.ErrNotFound
	}
	delete(m.st.MFATOTPs, userID)
	return nil
}

func (m *pgRepositoryMock) ListWebAuthnCredentials(_ context.Context, userID string) ([]models.WebAuthnCredential, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	var out []models.WebAuthnCredential
	for _, c := range m.st.WebAuthnCreds {
		if c.UserID == userID {
			out = append(out, c)
		}
	}
	return out, nil
}

func (m *pgRepositoryMock) InsertWebAuthnCredential(_ context.Context, cred models.WebAuthnCredential) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	m.st.WebAuthnCreds = append(m.st.WebAuthnCreds, cred)
	return nil
}

func (m *pgRepositoryMock) UpdateWebAuthnCredentialUsage(_ context.Context, credentialID []byte, signCount uint32, cloneWarning, backupState bool, now time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i := range m.st.WebAuthnCreds {
		if bytes.Equal(m.st.WebAuthnCreds[i].CredentialID, credentialID) {
			m.st.WebAuthnCreds[i].SignCount = signCount
			m.st.WebAuthnCreds[i].CloneWarning = cloneWarning
			m.st.WebAuthnCreds[i].BackupState = backupState
			t := now
			m.st.WebAuthnCreds[i].LastUsedAt = &t
			return nil
		}
	}
	return nil
}

func (m *pgRepositoryMock) RenameWebAuthnCredential(_ context.Context, userID, id, name string) (models.WebAuthnCredential, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i := range m.st.WebAuthnCreds {
		if m.st.WebAuthnCreds[i].UserID == userID && m.st.WebAuthnCreds[i].ID == id {
			m.st.WebAuthnCreds[i].Name = name
			return m.st.WebAuthnCreds[i], nil
		}
	}
	return models.WebAuthnCredential{}, constants.ErrNotFound
}

func (m *pgRepositoryMock) DeleteWebAuthnCredential(_ context.Context, userID, id string) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i := range m.st.WebAuthnCreds {
		if m.st.WebAuthnCreds[i].UserID == userID && m.st.WebAuthnCreds[i].ID == id {
			m.st.WebAuthnCreds = append(m.st.WebAuthnCreds[:i], m.st.WebAuthnCreds[i+1:]...)
			return nil
		}
	}
	return constants.ErrNotFound
}

func (m *pgRepositoryMock) CountUnusedRecoveryCodes(_ context.Context, userID string) (int32, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	var count int32
	for _, rc := range m.st.RecoveryCodes[userID] {
		if !rc.used {
			count++
		}
	}
	return count, nil
}

func (m *pgRepositoryMock) ReplaceRecoveryCodes(_ context.Context, userID string, codeHashes [][]byte, _ time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	if m.st.RecoveryCodes == nil {
		m.st.RecoveryCodes = make(map[string][]recoveryCodeMock)
	}
	rows := make([]recoveryCodeMock, len(codeHashes))
	for i, h := range codeHashes {
		rows[i] = recoveryCodeMock{hash: h}
	}
	m.st.RecoveryCodes[userID] = rows
	return nil
}

func (m *pgRepositoryMock) ConsumeRecoveryCode(_ context.Context, userID string, codeHash []byte, _ time.Time) (bool, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	rows := m.st.RecoveryCodes[userID]
	for i := range rows {
		if !rows[i].used && bytes.Equal(rows[i].hash, codeHash) {
			rows[i].used = true
			return true, nil
		}
	}
	return false, nil
}

func (m *pgRepositoryMock) ListRelationships(_ context.Context, userID string) ([]models.UserRelationship, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	var out []models.UserRelationship
	for _, rel := range m.st.Relationships {
		if rel.UserID == userID {
			out = append(out, rel)
		}
	}

	return out, nil
}

func (m *pgRepositoryMock) UpsertRelationship(_ context.Context, userID, targetUserID, kind string, now time.Time) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for i := range m.st.Relationships {
		if m.st.Relationships[i].UserID == userID && m.st.Relationships[i].TargetUserID == targetUserID {
			m.st.Relationships[i].Kind = kind
			m.st.Relationships[i].CreatedAt = now
			return nil
		}
	}
	m.st.Relationships = append(m.st.Relationships, models.UserRelationship{
		UserID:       userID,
		TargetUserID: targetUserID,
		Kind:         kind,
		CreatedAt:    now,
	})

	return nil
}

func (m *pgRepositoryMock) DeleteRelationship(_ context.Context, userID, targetUserID string) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	filtered := m.st.Relationships[:0]
	for _, rel := range m.st.Relationships {
		if rel.UserID == userID && rel.TargetUserID == targetUserID {
			continue
		}
		filtered = append(filtered, rel)
	}
	m.st.Relationships = filtered

	return nil
}
