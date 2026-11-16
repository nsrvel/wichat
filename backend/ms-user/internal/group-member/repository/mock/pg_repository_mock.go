package mock

import (
	"context"
	"sort"
	"strings"
	"time"

	groupmemberrepo "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

type pgRepositoryMock struct {
	st *usermock.State
}

// NewPgRepositoryMock is an in-memory group-member PgRepository backed by shared user mock state.
func NewPgRepositoryMock(st *usermock.State) groupmemberrepo.PgRepository {
	if st == nil {
		st = usermock.NewState()
	}
	return pgRepositoryMock{st: st}
}

func (m pgRepositoryMock) FindGroupMember(_ context.Context, groupID, userID string) (models.GroupMember, error) {
	m.st.Lock()
	defer m.st.Unlock()

	for _, mem := range m.st.Members {
		if mem.GroupID == groupID && mem.UserID == userID {
			return mem, nil
		}
	}
	return models.GroupMember{}, constants.ErrNotFound
}

func (m pgRepositoryMock) InsertGroupMember(_ context.Context, member models.GroupMember) error {
	m.st.Lock()
	defer m.st.Unlock()

	for _, mem := range m.st.Members {
		if mem.GroupID == member.GroupID && mem.UserID == member.UserID {
			return constants.ErrConflict
		}
	}
	m.st.Members = append(m.st.Members, member)
	return nil
}

func (m pgRepositoryMock) DeleteGroupMember(_ context.Context, groupID, userID string) error {
	m.st.Lock()
	defer m.st.Unlock()

	for i, mem := range m.st.Members {
		if mem.GroupID == groupID && mem.UserID == userID {
			m.st.Members = append(m.st.Members[:i], m.st.Members[i+1:]...)
			return nil
		}
	}
	return constants.ErrNotFound
}

func (m pgRepositoryMock) CountGroupMembers(_ context.Context, groupID string) (int, error) {
	m.st.Lock()
	defer m.st.Unlock()

	n := 0
	for _, mem := range m.st.Members {
		if mem.GroupID == groupID {
			n++
		}
	}
	return n, nil
}

func (m pgRepositoryMock) ListGroupMemberIDs(_ context.Context, groupID string) ([]string, error) {
	m.st.Lock()
	defer m.st.Unlock()

	var ids []string
	for _, mem := range m.st.Members {
		if mem.GroupID == groupID {
			ids = append(ids, mem.UserID)
		}
	}
	return ids, nil
}

func (m pgRepositoryMock) ListGroupMembersPage(
	_ context.Context,
	groupID string,
	limit int,
	cursor *models.GroupMemberCursor,
) ([]models.GroupMemberProfile, *models.GroupMemberCursor, error) {
	m.st.Lock()
	defer m.st.Unlock()

	if limit <= 0 {
		limit = constants.DefaultMemberListLimit
	}
	if limit > constants.MaxMemberListLimit {
		limit = constants.MaxMemberListLimit
	}

	var list []models.GroupMember
	for _, mem := range m.st.Members {
		if mem.GroupID == groupID {
			list = append(list, mem)
		}
	}

	sort.Slice(list, func(i, j int) bool {
		if !list[i].CreatedAt.Equal(list[j].CreatedAt) {
			return list[i].CreatedAt.Before(list[j].CreatedAt)
		}
		return list[i].UserID < list[j].UserID
	})

	var filtered []models.GroupMember
	for _, mem := range list {
		if cursor != nil {
			if mem.CreatedAt.Before(cursor.CreatedAt) || (mem.CreatedAt.Equal(cursor.CreatedAt) && mem.UserID <= cursor.UserID) {
				continue
			}
		}
		filtered = append(filtered, mem)
	}

	var next *models.GroupMemberCursor
	if len(filtered) > limit {
		last := filtered[limit-1]
		next = &models.GroupMemberCursor{CreatedAt: last.CreatedAt, UserID: last.UserID}
		filtered = filtered[:limit]
	}

	userMap := make(map[string]models.User, len(m.st.Users))
	for _, u := range m.st.Users {
		userMap[u.ID] = u
	}

	out := make([]models.GroupMemberProfile, len(filtered))
	for i, mem := range filtered {
		u := userMap[mem.UserID]
		displayName := u.DisplayName
		if displayName == "" {
			displayName = mem.UserID
		}
		joinMethod := mem.JoinMethod
		if joinMethod == "" {
			joinMethod = constants.JoinMethodUnknown
		}
		var invitedByID, invitedByUsername, invitedByDisplayName string
		if mem.InvitedByUserID != nil && *mem.InvitedByUserID != "" {
			invitedByID = *mem.InvitedByUserID
			if inv, ok := userMap[invitedByID]; ok {
				invitedByUsername = inv.Username
				invitedByDisplayName = inv.DisplayName
			}
		}
		out[i] = models.GroupMemberProfile{
			UserID:               mem.UserID,
			DisplayName:          displayName,
			Username:             u.Username,
			AvatarURL:            u.AvatarURL,
			BannerColor:          u.BannerColor,
			Bio:                  u.Bio,
			Pronouns:             u.Pronouns,
			UserCreatedAt:        u.CreatedAt,
			StatusNoteText:       u.StatusNoteText,
			StatusNoteEmoji:      u.StatusNoteEmoji,
			StatusNoteExpiresAt:  u.StatusNoteExpiresAt,
			JoinMethod:           joinMethod,
			InviteCode:           mem.InviteCode,
			InvitedByUserID:      invitedByID,
			InvitedByUsername:    invitedByUsername,
			InvitedByDisplayName: invitedByDisplayName,
			TimeoutUntil:         mem.TimeoutUntil,
			Role:                 mem.Role,
			JoinedAt:             mem.CreatedAt,
		}
	}
	return out, next, nil
}

func (m pgRepositoryMock) IsMemberBanned(_ context.Context, groupID, userID string) (bool, error) {
	m.st.Lock()
	defer m.st.Unlock()

	now := time.Now().UTC()
	for _, b := range m.st.Bans {
		if b.GroupID == groupID && b.UserID == userID && (b.ExpiresAt == nil || b.ExpiresAt.After(now)) {
			return true, nil
		}
	}
	return false, nil
}

func (m pgRepositoryMock) BanMembers(_ context.Context, groupID string, userIDs []string, bannedBy, reason string, deleteSeconds int, expiresAt *time.Time) (int, error) {
	m.st.Lock()
	defer m.st.Unlock()

	count := 0
	for _, uid := range userIDs {
		isOwner := false
		for _, mem := range m.st.Members {
			if mem.GroupID == groupID && mem.UserID == uid && mem.Role == constants.GroupRoleOwner {
				isOwner = true
				break
			}
		}
		if isOwner {
			continue
		}

		var bBy *string
		if bannedBy != "" {
			bBy = &bannedBy
		}
		m.st.Bans = append(m.st.Bans, models.GroupBanRow{
			GroupID: groupID, UserID: uid, BannedBy: bBy, Reason: reason, DeleteMessageSeconds: deleteSeconds, ExpiresAt: expiresAt, CreatedAt: time.Now().UTC(),
		})

		var filteredMembers []models.GroupMember
		for _, mem := range m.st.Members {
			if mem.GroupID != groupID || mem.UserID != uid {
				filteredMembers = append(filteredMembers, mem)
			}
		}
		m.st.Members = filteredMembers

		var filteredRoles []models.GroupMemberRoleRow
		for _, mr := range m.st.MemberRoles {
			if mr.GroupID != groupID || mr.UserID != uid {
				filteredRoles = append(filteredRoles, mr)
			}
		}
		m.st.MemberRoles = filteredRoles

		count++
	}
	return count, nil
}

func (m pgRepositoryMock) ListBansPage(
	_ context.Context,
	groupID, search string,
	limit int,
	cursor *models.GroupMemberCursor,
) ([]models.GroupBanEntry, *models.GroupMemberCursor, int, error) {
	m.st.Lock()
	defer m.st.Unlock()

	if limit <= 0 {
		limit = constants.DefaultMemberListLimit
	}
	now := time.Now().UTC()
	q := strings.ToLower(strings.TrimSpace(search))

	userByID := make(map[string]models.User, len(m.st.Users))
	for _, u := range m.st.Users {
		userByID[u.ID] = u
	}

	var all []models.GroupBanEntry
	for _, b := range m.st.Bans {
		if b.GroupID != groupID {
			continue
		}
		if b.ExpiresAt != nil && !b.ExpiresAt.After(now) {
			continue
		}
		u, ok := userByID[b.UserID]
		if !ok {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(u.Username), q) &&
			!strings.Contains(strings.ToLower(u.DisplayName), q) &&
			!strings.Contains(strings.ToLower(b.Reason), q) {
			continue
		}
		var bannedByID, bannedByUsername, bannedByDisplayName string
		if b.BannedBy != nil {
			bannedByID = *b.BannedBy
			if bu, ok := userByID[bannedByID]; ok {
				bannedByUsername = bu.Username
				bannedByDisplayName = bu.DisplayName
			}
		}
		all = append(all, models.GroupBanEntry{
			UserID:              u.ID,
			Username:            u.Username,
			DisplayName:         u.DisplayName,
			AvatarURL:           u.AvatarURL,
			Reason:              b.Reason,
			BannedByUserID:      bannedByID,
			BannedByUsername:    bannedByUsername,
			BannedByDisplayName: bannedByDisplayName,
			ExpiresAt:           b.ExpiresAt,
			CreatedAt:           b.CreatedAt,
		})
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].CreatedAt.Equal(all[j].CreatedAt) {
			return all[i].UserID < all[j].UserID
		}
		return all[i].CreatedAt.After(all[j].CreatedAt)
	})
	total := len(all)

	var filtered []models.GroupBanEntry
	for _, item := range all {
		if cursor != nil {
			if item.CreatedAt.After(cursor.CreatedAt) || (item.CreatedAt.Equal(cursor.CreatedAt) && item.UserID <= cursor.UserID) {
				continue
			}
		}
		filtered = append(filtered, item)
	}

	var next *models.GroupMemberCursor
	if len(filtered) > limit {
		last := filtered[limit-1]
		next = &models.GroupMemberCursor{CreatedAt: last.CreatedAt, UserID: last.UserID}
		filtered = filtered[:limit]
	}
	return filtered, next, total, nil
}

func (m pgRepositoryMock) UnbanMember(_ context.Context, groupID, userID string) error {
	m.st.Lock()
	defer m.st.Unlock()

	for i, b := range m.st.Bans {
		if b.GroupID == groupID && b.UserID == userID {
			m.st.Bans = append(m.st.Bans[:i], m.st.Bans[i+1:]...)
			return nil
		}
	}
	return constants.ErrNotFound
}

func (m pgRepositoryMock) UpdateGroupMemberTimeout(_ context.Context, groupID, userID string, timeoutUntil *time.Time) error {
	m.st.Lock()
	defer m.st.Unlock()

	for i := range m.st.Members {
		if m.st.Members[i].GroupID == groupID && m.st.Members[i].UserID == userID {
			m.st.Members[i].TimeoutUntil = timeoutUntil
			return nil
		}
	}
	return constants.ErrNotFound
}
