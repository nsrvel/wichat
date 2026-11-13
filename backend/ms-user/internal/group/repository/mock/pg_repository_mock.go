package mock

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

type pgRepositoryMock struct {
	st *usermock.State
}

// NewPgRepositoryMock is an in-memory group PgRepository backed by shared user mock state.
func NewPgRepositoryMock(st *usermock.State) repository.PgRepository {
	return pgRepositoryMock{st: st}
}

func (m pgRepositoryMock) GetFirstGroup(context.Context) (models.Group, error) {
	m.st.Lock()
	defer m.st.Unlock()

	if len(m.st.Groups) == 0 {
		return models.Group{}, constants.ErrNotFound
	}

	return m.st.Groups[0], nil
}

func (m pgRepositoryMock) ListGroupsForUser(_ context.Context, userID string) ([]models.Group, error) {
	m.st.Lock()
	defer m.st.Unlock()

	return groupsForUser(m.st, userID), nil
}

func (m pgRepositoryMock) ListJoinableGroups(_ context.Context, userID string) ([]models.Group, error) {
	m.st.Lock()
	defer m.st.Unlock()

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

func (m pgRepositoryMock) FindGroupByID(_ context.Context, id string) (models.Group, error) {
	m.st.Lock()
	defer m.st.Unlock()

	for _, g := range m.st.Groups {
		if g.ID == id {
			return g, nil
		}
	}

	return models.Group{}, constants.ErrNotFound
}

func (m pgRepositoryMock) FindGroupByPublicID(_ context.Context, publicID string) (models.Group, error) {
	m.st.Lock()
	defer m.st.Unlock()

	for _, g := range m.st.Groups {
		if g.PublicID == publicID {
			return g, nil
		}
	}

	return models.Group{}, constants.ErrNotFound
}

func (m pgRepositoryMock) CreateGroupWithOwner(_ context.Context, group models.Group, owner models.GroupMember) error {
	m.st.Lock()
	defer m.st.Unlock()

	m.st.Groups = append(m.st.Groups, group)
	m.st.Members = append(m.st.Members, owner)

	everyone := models.GroupRole{
		ID:          group.ID + "-everyone",
		PublicID:    strings.Repeat("e", 24),
		GroupID:     group.ID,
		Slug:        constants.RoleSlugEveryone,
		Name:        "@everyone",
		Color:       "#99aab5",
		IsDefault:   true,
		IsSystem:    true,
		Position:    0,
		Permissions: constants.DefaultEveryonePermissions,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}
	member := models.GroupRole{
		ID:          group.ID + "-member",
		PublicID:    strings.Repeat("m", 24),
		GroupID:     group.ID,
		Slug:        constants.RoleSlugMember,
		Name:        "Member",
		Color:       "#5865f2",
		IsDefault:   false,
		IsSystem:    true,
		Position:    1,
		Permissions: []string{},
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}
	m.st.Roles = append(m.st.Roles, everyone, member)
	m.st.MemberRoles = append(m.st.MemberRoles, models.GroupMemberRoleRow{
		GroupID:   group.ID,
		UserID:    owner.UserID,
		RoleID:    member.ID,
		CreatedAt: group.CreatedAt,
	})

	return nil
}

func (m pgRepositoryMock) DeleteGroup(_ context.Context, groupID string) error {
	m.st.Lock()
	defer m.st.Unlock()

	found := false
	filtered := m.st.Groups[:0]
	for _, g := range m.st.Groups {
		if g.ID == groupID {
			found = true
			continue
		}
		filtered = append(filtered, g)
	}
	if !found {
		return constants.ErrNotFound
	}
	m.st.Groups = filtered

	var members []models.GroupMember
	for _, mem := range m.st.Members {
		if mem.GroupID != groupID {
			members = append(members, mem)
		}
	}
	m.st.Members = members

	return nil
}

func (m pgRepositoryMock) UpdateGroupDisplayName(_ context.Context, groupID, name string, _ time.Time) error {
	m.st.Lock()
	defer m.st.Unlock()

	for i := range m.st.Groups {
		if m.st.Groups[i].ID == groupID {
			m.st.Groups[i].Name = name

			return nil
		}
	}

	return constants.ErrNotFound
}

func (m pgRepositoryMock) UpdateGroupLogoURL(_ context.Context, groupID, logoURL string, _ time.Time) error {
	m.st.Lock()
	defer m.st.Unlock()

	for i := range m.st.Groups {
		if m.st.Groups[i].ID == groupID {
			m.st.Groups[i].LogoURL = logoURL

			return nil
		}
	}

	return constants.ErrNotFound
}

func (m pgRepositoryMock) UpdateGroupProfile(_ context.Context, groupID string, patch models.GroupProfilePatch, _ time.Time) error {
	m.st.Lock()
	defer m.st.Unlock()

	for i := range m.st.Groups {
		if m.st.Groups[i].ID == groupID {
			m.st.Groups[i].Name = patch.Name
			m.st.Groups[i].Description = patch.Description
			m.st.Groups[i].IsPrivate = patch.IsPrivate
			m.st.Groups[i].BannerColor = patch.BannerColor
			m.st.Groups[i].BannerURL = patch.BannerURL
			m.st.Groups[i].LogoURL = patch.LogoURL

			return nil
		}
	}

	return constants.ErrNotFound
}

func (m pgRepositoryMock) SetRailPositions(_ context.Context, userID string, orderedGroupIDs []string) error {
	m.st.Lock()
	defer m.st.Unlock()

	for i, groupID := range orderedGroupIDs {
		found := false
		for j := range m.st.Members {
			if m.st.Members[j].UserID == userID && m.st.Members[j].GroupID == groupID {
				pos := i
				m.st.Members[j].RailPosition = &pos
				found = true
				break
			}
		}
		if !found {
			return constants.ErrNotFound
		}
	}

	return nil
}

func groupsForUser(st *usermock.State, userID string) []models.Group {
	memberOf := map[string]string{}
	railPos := map[string]*int{}
	for _, mem := range st.Members {
		if mem.UserID == userID {
			memberOf[mem.GroupID] = mem.Role
			railPos[mem.GroupID] = mem.RailPosition
		}
	}
	var out []models.Group
	for _, g := range st.Groups {
		if role, ok := memberOf[g.ID]; ok {
			g.CallerRole = role
			out = append(out, g)
		}
	}

	sort.SliceStable(out, func(i, j int) bool {
		pi, pj := railPos[out[i].ID], railPos[out[j].ID]
		if pi == nil && pj == nil {
			return out[i].Name < out[j].Name
		}
		if pi == nil {
			return false
		}
		if pj == nil {
			return true
		}
		if *pi != *pj {
			return *pi < *pj
		}
		return out[i].Name < out[j].Name
	})

	return out
}
