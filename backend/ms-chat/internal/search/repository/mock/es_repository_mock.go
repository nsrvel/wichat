package mock

import (
	"context"
	"strings"
	"sync"

	"github.com/wichat/wichat/backend/ms-chat/internal/search/repository"
	"github.com/wichat/wichat/backend/wi-shared/infra/elasticsearch"
)

// State records ES index calls in tests.
type State struct {
	mu             sync.Mutex
	Ensured        int
	Docs           []elasticsearch.MessageDoc
	AttachmentDocs []elasticsearch.AttachmentDoc
	DeletedIDs     []string
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

func (m esRepositoryMock) DeleteMessage(_ context.Context, messageID string) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()
	m.st.DeletedIDs = append(m.st.DeletedIDs, messageID)

	return nil
}

func (m esRepositoryMock) SearchMessages(_ context.Context, groupID, query string, limit int) ([]elasticsearch.MessageDoc, error) {
	groupID = strings.TrimSpace(groupID)
	query = strings.TrimSpace(strings.ToLower(query))
	if groupID == "" || query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 8
	}

	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	out := make([]elasticsearch.MessageDoc, 0, limit)
	for _, doc := range m.st.Docs {
		if doc.GroupID != groupID {
			continue
		}
		if !strings.Contains(strings.ToLower(doc.Body), query) {
			continue
		}
		out = append(out, doc)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (m esRepositoryMock) IndexAttachmentCreated(_ context.Context, doc elasticsearch.AttachmentDoc) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()
	m.st.AttachmentDocs = append(m.st.AttachmentDocs, doc)
	return nil
}

func (m esRepositoryMock) SearchFiles(_ context.Context, groupID, query string, limit int, imagesOnly bool) ([]elasticsearch.AttachmentDoc, error) {
	groupID = strings.TrimSpace(groupID)
	query = strings.TrimSpace(strings.ToLower(query))
	if groupID == "" || query == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 8
	}

	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	out := make([]elasticsearch.AttachmentDoc, 0, limit)
	for _, doc := range m.st.AttachmentDocs {
		if doc.GroupID != groupID {
			continue
		}
		if imagesOnly && !strings.HasPrefix(strings.ToLower(doc.ContentType), "image/") {
			continue
		}
		if !strings.Contains(strings.ToLower(doc.Filename), query) {
			continue
		}
		out = append(out, doc)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}
