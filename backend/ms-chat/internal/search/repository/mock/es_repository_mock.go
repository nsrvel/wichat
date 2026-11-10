package mock

import (
	"context"
	"sync"

	"github.com/wichat/wichat/backend/ms-chat/internal/search/repository"
	"github.com/wichat/wichat/backend/wi-shared/infra/elasticsearch"
)

// State records ES index calls in tests.
type State struct {
	mu      sync.Mutex
	Ensured int
	Docs    []elasticsearch.MessageDoc
}

// NewState returns empty ES mock state.
func NewState() *State {
	return &State{}
}

type esRepositoryMock struct {
	st *State
}

// NewEsRepositoryMock is an in-memory EsRepository for service and consumer tests.
func NewEsRepositoryMock(st *State) repository.EsRepository {
	return esRepositoryMock{st: st}
}

func (m esRepositoryMock) EnsureIndex(_ context.Context) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()
	m.st.Ensured++

	return nil
}

func (m esRepositoryMock) IndexMessageCreated(_ context.Context, doc elasticsearch.MessageDoc) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()
	m.st.Docs = append(m.st.Docs, doc)

	return nil
}
