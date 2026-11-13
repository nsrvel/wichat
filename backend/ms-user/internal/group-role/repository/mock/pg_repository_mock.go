package mock

import (
	"context"
	"strings"
	"time"

	"github.com/uptrace/bun"
	grouprolerepo "github.com/wichat/wichat/backend/ms-user/internal/group-role/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

type pgRepositoryMock struct {
	st *usermock.State
}

// NewPgRepositoryMock is an in-memory group-role PgRepository backed by shared user mock state.
func NewPgRepositoryMock(st *usermock.State) grouprolerepo.PgRepository {
	if st == nil {
		st = usermock.NewState()
	}
	return pgRepositoryMock{st: st}
}

func (m pgRepositoryMock) GetEffectivePermissions(_ context.Context, groupID, userID string) ([]string, error) {
	m.st.Lock()
	defer m.st.Unlock()

	isOwner := false
	for _, mem := range m.st.Members {
		if mem.GroupID == groupID && mem.UserID == userID {
			if mem.Role == constants.GroupRoleOwner {
				isOwner = true
			}
			break
		}
	}
	if isOwner {
		return append([]string(nil), constants.AllKnownPermissions...), nil
	}

	var everyonePerms []string
	for _, r := range m.st.Roles {
		if r.GroupID == groupID && r.IsDefault {
			everyonePerms = r.Permissions
			break
		}
	}

	var rolePerms [][]string
	for _, mr := range m.st.MemberRoles {
		if mr.GroupID == groupID && mr.UserID == userID {
			for _, r := range m.st.Roles {
				if r.ID == mr.RoleID {
					rolePerms = append(rolePerms, r.Permissions)
				}
			}
		}
	}

	set := make(map[string]struct{})
	for _, p := range everyonePerms {
		set[p] = struct{}{}
	}
	for _, rp := range rolePerms {
		for _, p := range rp {
			set[p] = struct{}{}
		}
	}

	if _, ok := set[constants.PermAdministrator]; ok {
		return append([]string(nil), constants.AllKnownPermissions...), nil
	}

	out := make([]string, 0, len(set))
	for _, p := range constants.AllKnownPermissions {
		if _, ok := set[p]; ok {
			out = append(out, p)
		}
	}
	return out, nil
}

func (m pgRepositoryMock) ListRolesForGroup(_ context.Context, groupID string) ([]models.GroupRole, error) {
	m.st.Lock()
	defer m.st.Unlock()

	var out []models.GroupRole
	for _, r := range m.st.Roles {
		if r.GroupID == groupID {
			roleCopy := r
			if r.IsDefault {
				count := 0
				for _, mem := range m.st.Members {
					if mem.GroupID == groupID {
						count++
					}
				}
				roleCopy.MemberCount = count
			} else {
				count := 0
				for _, mr := range m.st.MemberRoles {
					if mr.RoleID == r.ID {
						count++
					}
				}
				roleCopy.MemberCount = count
			}
			out = append(out, roleCopy)
		}
	}
	return out, nil
}

func (m pgRepositoryMock) FindRoleByID(_ context.Context, groupID, roleID string) (models.GroupRole, error) {
	m.st.Lock()
	defer m.st.Unlock()

	for _, r := range m.st.Roles {
		if r.GroupID == groupID && (r.ID == roleID || r.PublicID == roleID) {
			return r, nil
		}
	}
	return models.GroupRole{}, constants.ErrNotFound
}

func (m pgRepositoryMock) UpdateRole(_ context.Context, groupID, roleID string, patch models.GroupRolePatch, now time.Time) (models.GroupRole, error) {
	m.st.Lock()
	defer m.st.Unlock()

	for i := range m.st.Roles {
		r := &m.st.Roles[i]
		if r.GroupID == groupID && (r.ID == roleID || r.PublicID == roleID) {
			if patch.Permissions != nil {
				r.Permissions = patch.Permissions
			}
			if !r.IsDefault {
				if patch.Name != nil {
					r.Name = *patch.Name
				}
				if patch.Color != nil {
					r.Color = *patch.Color
				}
				if patch.Hoist != nil {
					r.Hoist = *patch.Hoist
				}
				if patch.Mentionable != nil {
					r.Mentionable = *patch.Mentionable
				}
			}
			r.UpdatedAt = now
			return *r, nil
		}
	}
	return models.GroupRole{}, constants.ErrNotFound
}

func (m pgRepositoryMock) AddMemberRoles(_ context.Context, groupID, roleID string, userIDs []string) error {
	m.st.Lock()
	defer m.st.Unlock()

	var resolvedID string
	for _, r := range m.st.Roles {
		if r.GroupID == groupID && (r.ID == roleID || r.PublicID == roleID) {
			if r.IsDefault {
				return nil
			}
			resolvedID = r.ID
			break
		}
	}
	if resolvedID == "" {
		return constants.ErrNotFound
	}

	for _, uid := range userIDs {
		exists := false
		for _, mem := range m.st.Members {
			if mem.GroupID == groupID && mem.UserID == uid {
				exists = true
				break
			}
		}
		if !exists {
			continue
		}
		alreadyAssigned := false
		for _, mr := range m.st.MemberRoles {
			if mr.GroupID == groupID && mr.UserID == uid && mr.RoleID == resolvedID {
				alreadyAssigned = true
				break
			}
		}
		if !alreadyAssigned {
			m.st.MemberRoles = append(m.st.MemberRoles, models.GroupMemberRoleRow{
				GroupID: groupID, UserID: uid, RoleID: resolvedID, CreatedAt: time.Now().UTC(),
			})
		}
	}
	return nil
}

func (m pgRepositoryMock) RemoveMemberRole(_ context.Context, groupID, roleID, userID string) error {
	m.st.Lock()
	defer m.st.Unlock()

	var resolvedID string
	for _, r := range m.st.Roles {
		if r.GroupID == groupID && (r.ID == roleID || r.PublicID == roleID) {
			resolvedID = r.ID
			break
		}
	}
	if resolvedID == "" {
		return constants.ErrNotFound
	}

	filtered := m.st.MemberRoles[:0]
	removed := false
	for _, mr := range m.st.MemberRoles {
		if mr.GroupID == groupID && mr.RoleID == resolvedID && mr.UserID == userID {
			removed = true
			continue
		}
		filtered = append(filtered, mr)
	}
	m.st.MemberRoles = filtered
	if !removed {
		return constants.ErrNotFound
	}
	return nil
}

func (m pgRepositoryMock) ListRolesForMembers(_ context.Context, groupID string, userIDs []string) (map[string][]models.GroupRoleSummary, error) {
	m.st.Lock()
	defer m.st.Unlock()

	out := make(map[string][]models.GroupRoleSummary, len(userIDs))
	for _, uid := range userIDs {
		var summaries []models.GroupRoleSummary
		for _, mr := range m.st.MemberRoles {
			if mr.GroupID == groupID && mr.UserID == uid {
				for _, r := range m.st.Roles {
					if r.ID == mr.RoleID {
						summaries = append(summaries, models.GroupRoleSummary{
							ID: r.PublicID, Slug: r.Slug, Name: r.Name, Color: r.Color,
						})
					}
				}
			}
		}
		out[uid] = summaries
	}
	return out, nil
}

func (m pgRepositoryMock) SeedDefaultGroupRolesTx(_ context.Context, _ bun.Tx, groupID, ownerUserID string, now time.Time) error {
	m.st.Lock()
	defer m.st.Unlock()

	everyone := models.GroupRole{
		ID:          groupID + "-everyone",
		PublicID:    strings.Repeat("e", 24),
		GroupID:     groupID,
		Slug:        constants.RoleSlugEveryone,
		Name:        "@everyone",
		Color:       "#99aab5",
		IsDefault:   true,
		IsSystem:    true,
		Position:    0,
		Permissions: constants.DefaultEveryonePermissions,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	member := models.GroupRole{
		ID:          groupID + "-member",
		PublicID:    strings.Repeat("m", 24),
		GroupID:     groupID,
		Slug:        constants.RoleSlugMember,
		Name:        "Member",
		Color:       "#5865f2",
		IsDefault:   false,
		IsSystem:    true,
		Position:    1,
		Permissions: []string{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	m.st.Roles = append(m.st.Roles, everyone, member)
	m.st.MemberRoles = append(m.st.MemberRoles, models.GroupMemberRoleRow{
		GroupID:   groupID,
		UserID:    ownerUserID,
		RoleID:    member.ID,
		CreatedAt: now,
	})
	return nil
}
