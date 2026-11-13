package mock

import (
	"context"
	"sort"
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
		out[i] = models.GroupMemberProfile{
			UserID:      mem.UserID,
			DisplayName: displayName,
			Email:       u.Email,
			Role:        mem.Role,
			JoinedAt:    mem.CreatedAt,
		}
	}
	return out, next, nil
}

func (m pgRepositoryMock) IsMemberBanned(_ context.Context, groupID, userID string) (bool, error) {
	m.st.Lock()
	defer m.st.Unlock()

	for _, b := range m.st.Bans {
		if b.GroupID == groupID && b.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (m pgRepositoryMock) BanMembers(_ context.Context, groupID string, userIDs []string, bannedBy, reason string, deleteSeconds int) (int, error) {
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
			GroupID: groupID, UserID: uid, BannedBy: bBy, Reason: reason, DeleteMessageSeconds: deleteSeconds, CreatedAt: time.Now().UTC(),
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
