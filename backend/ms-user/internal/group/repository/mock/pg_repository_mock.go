package mock

import (
	"context"
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

func (m pgRepositoryMock) CreateGroupWithOwner(_ context.Context, group models.Group, owner models.GroupMember) error {
	m.st.Lock()
	defer m.st.Unlock()

	m.st.Groups = append(m.st.Groups, group)
	m.st.Members = append(m.st.Members, owner)

	return nil
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

func groupsForUser(st *usermock.State, userID string) []models.Group {
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
