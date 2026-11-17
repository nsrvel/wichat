package mock

import (
	"context"
	"sync"
	"time"

	"github.com/wichat/wichat/backend/ms-core/internal/models"
	"github.com/wichat/wichat/backend/ms-core/internal/voice/repository"
)

// State stores in-memory call sessions, participants, and recordings for unit tests.
type State struct {
	mu           sync.Mutex
	Sessions     map[string]models.CallSession
	Participants map[string][]models.CallParticipant
	Recordings   map[string]models.CallRecording
}

// NewState constructs an empty mock repository state.
func NewState() *State {
	return &State{
		Sessions:     make(map[string]models.CallSession),
		Participants: make(map[string][]models.CallParticipant),
		Recordings:   make(map[string]models.CallRecording),
	}
}

type pgRepositoryMock struct {
	st *State
}

// NewPgRepositoryMock builds an in-memory PgRepository.
func NewPgRepositoryMock(st *State) repository.PgRepository {
	return pgRepositoryMock{st: st}
}

func (m pgRepositoryMock) UpsertActiveSession(_ context.Context, session models.CallSession) (models.CallSession, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for _, existing := range m.st.Sessions {
		if existing.ChannelID == session.ChannelID && existing.EndedAt == nil {
			return existing, nil
		}
	}
	m.st.Sessions[session.ID] = session
	return session, nil
}

func (m pgRepositoryMock) FindActiveSessionByChannel(_ context.Context, channelID string) (models.CallSession, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for _, s := range m.st.Sessions {
		if s.ChannelID == channelID && s.EndedAt == nil {
			return s, nil
		}
	}
	return models.CallSession{}, repository.ErrNotFound
}

func (m pgRepositoryMock) UpdateSessionPeak(_ context.Context, sessionID string, peak int32) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	s, ok := m.st.Sessions[sessionID]
	if !ok {
		return repository.ErrNotFound
	}
	if peak > s.PeakParticipants {
		s.PeakParticipants = peak
		m.st.Sessions[sessionID] = s
	}
	return nil
}

func (m pgRepositoryMock) CloseSession(_ context.Context, sessionID string, endedAt time.Time, addParticipantSeconds int64) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	s, ok := m.st.Sessions[sessionID]
	if !ok {
		return repository.ErrNotFound
	}
	if s.EndedAt == nil {
		s.EndedAt = &endedAt
		s.TotalParticipantSeconds += addParticipantSeconds
		m.st.Sessions[sessionID] = s
	}

	parts := m.st.Participants[sessionID]
	for i := range parts {
		if parts[i].LeftAt == nil {
			parts[i].LeftAt = &endedAt
			parts[i].DurationSeconds = int32(max(0, int(endedAt.Sub(parts[i].JoinedAt).Seconds())))
			if parts[i].DisconnectReason == "" {
				parts[i].DisconnectReason = "room_closed"
			}
		}
	}
	m.st.Participants[sessionID] = parts
	return nil
}

func (m pgRepositoryMock) InsertParticipantJoin(_ context.Context, participant models.CallParticipant) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	m.st.Participants[participant.SessionID] = append(m.st.Participants[participant.SessionID], participant)
	return nil
}

func (m pgRepositoryMock) CloseParticipantLeave(_ context.Context, sessionID, userID string, leftAt time.Time, reason string) (int32, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	parts := m.st.Participants[sessionID]
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i].UserID == userID && parts[i].LeftAt == nil {
			dur := int32(max(0, int(leftAt.Sub(parts[i].JoinedAt).Seconds())))
			parts[i].LeftAt = &leftAt
			parts[i].DurationSeconds = dur
			parts[i].DisconnectReason = reason
			m.st.Participants[sessionID] = parts
			if s, ok := m.st.Sessions[sessionID]; ok {
				s.TotalParticipantSeconds += int64(dur)
				m.st.Sessions[sessionID] = s
			}
			return dur, nil
		}
	}
	return 0, nil
}

func (m pgRepositoryMock) UpdateParticipantModeration(_ context.Context, sessionID, userID string, serverMuted, serverDeafened *bool) error {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	parts := m.st.Participants[sessionID]
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i].UserID == userID && parts[i].LeftAt == nil {
			if serverMuted != nil {
				parts[i].ServerMuted = *serverMuted
			}
			if serverDeafened != nil {
				parts[i].ServerDeafened = *serverDeafened
			}
		}
	}
	m.st.Participants[sessionID] = parts
	return nil
}

func (m pgRepositoryMock) CreateRecording(_ context.Context, rec models.CallRecording) (models.CallRecording, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	m.st.Recordings[rec.EgressID] = rec
	return rec, nil
}

func (m pgRepositoryMock) FindRecordingByEgressID(_ context.Context, egressID string) (models.CallRecording, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	rec, ok := m.st.Recordings[egressID]
	if !ok {
		return models.CallRecording{}, repository.ErrNotFound
	}
	return rec, nil
}

func (m pgRepositoryMock) FindActiveRecordingByChannel(_ context.Context, channelID string) (models.CallRecording, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	for _, rec := range m.st.Recordings {
		if rec.ChannelID == channelID && (rec.Status == "starting" || rec.Status == "active" || rec.Status == "stopping") {
			return rec, nil
		}
	}
	return models.CallRecording{}, repository.ErrNotFound
}

func (m pgRepositoryMock) UpdateRecordingByEgressID(
	_ context.Context,
	egressID, status, fileKey string,
	fileSizeBytes int64,
	durationSeconds int32,
	stoppedBy *string,
	endedAt *time.Time,
) (models.CallRecording, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	rec, ok := m.st.Recordings[egressID]
	if !ok {
		return models.CallRecording{}, repository.ErrNotFound
	}
	rec.Status = status
	if fileKey != "" {
		rec.FileKey = fileKey
	}
	if fileSizeBytes > 0 {
		rec.FileSizeBytes = fileSizeBytes
	}
	if durationSeconds > 0 {
		rec.DurationSeconds = durationSeconds
	}
	if stoppedBy != nil {
		rec.StoppedBy = stoppedBy
	}
	if endedAt != nil {
		rec.EndedAt = endedAt
	}
	m.st.Recordings[egressID] = rec
	return rec, nil
}

func (m pgRepositoryMock) ListRecordingsByChannel(_ context.Context, channelID string, limit int) ([]models.CallRecording, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	var out []models.CallRecording
	for _, rec := range m.st.Recordings {
		if rec.ChannelID == channelID {
			out = append(out, rec)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (m pgRepositoryMock) SumParticipantSecondsBetween(_ context.Context, from, to time.Time) (int64, error) {
	m.st.mu.Lock()
	defer m.st.mu.Unlock()

	var total int64
	for _, s := range m.st.Sessions {
		if !s.StartedAt.Before(from) && s.StartedAt.Before(to) {
			total += s.TotalParticipantSeconds
		}
	}
	return total, nil
}
