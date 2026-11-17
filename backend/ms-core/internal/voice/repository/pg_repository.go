package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/uptrace/bun"
	"github.com/wichat/wichat/backend/ms-core/internal/models"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// ErrNotFound indicates a missing record.
var ErrNotFound = errors.New("not found")

// PgRepository persists call sessions, participants, and recordings in PostgreSQL.
type PgRepository interface {
	UpsertActiveSession(ctx context.Context, session models.CallSession) (models.CallSession, error)
	FindActiveSessionByChannel(ctx context.Context, channelID string) (models.CallSession, error)
	UpdateSessionPeak(ctx context.Context, sessionID string, peak int32) error
	CloseSession(ctx context.Context, sessionID string, endedAt time.Time, addParticipantSeconds int64) error
	InsertParticipantJoin(ctx context.Context, participant models.CallParticipant) error
	CloseParticipantLeave(ctx context.Context, sessionID, userID string, leftAt time.Time, reason string) (int32, error)
	UpdateParticipantModeration(ctx context.Context, sessionID, userID string, serverMuted, serverDeafened *bool) error
	CreateRecording(ctx context.Context, rec models.CallRecording) (models.CallRecording, error)
	FindRecordingByEgressID(ctx context.Context, egressID string) (models.CallRecording, error)
	FindActiveRecordingByChannel(ctx context.Context, channelID string) (models.CallRecording, error)
	UpdateRecordingByEgressID(ctx context.Context, egressID, status, fileKey string, fileSizeBytes int64, durationSeconds int32, stoppedBy *string, endedAt *time.Time) (models.CallRecording, error)
	ListRecordingsByChannel(ctx context.Context, channelID string, limit int) ([]models.CallRecording, error)
	SumParticipantSecondsBetween(ctx context.Context, from, to time.Time) (int64, error)
}

type pgRepository struct {
	db  *bun.DB
	log logger.Logger
}

// NewPgRepository constructs a Postgres voice store.
func NewPgRepository(db *bun.DB, log logger.Logger) PgRepository {
	return pgRepository{db: db, log: log}
}

func (r pgRepository) UpsertActiveSession(ctx context.Context, session models.CallSession) (models.CallSession, error) {
	existing, err := r.FindActiveSessionByChannel(ctx, session.ChannelID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return models.CallSession{}, err
	}

	row := models.CallSessionRow{
		ID:                      session.ID,
		PublicID:                session.PublicID,
		GroupID:                 session.GroupID,
		ChannelID:               session.ChannelID,
		RoomName:                session.RoomName,
		StartedAt:               session.StartedAt,
		PeakParticipants:        session.PeakParticipants,
		TotalParticipantSeconds: session.TotalParticipantSeconds,
	}
	if _, err := r.db.NewInsert().Model(&row).Exec(ctx); err != nil {
		return models.CallSession{}, err
	}

	return row.ToModel(), nil
}

func (r pgRepository) FindActiveSessionByChannel(ctx context.Context, channelID string) (models.CallSession, error) {
	var row models.CallSessionRow
	err := r.db.NewSelect().Model(&row).
		Where("cs.channel_id = ? AND cs.ended_at IS NULL", channelID).
		OrderExpr("cs.started_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CallSession{}, ErrNotFound
		}
		return models.CallSession{}, err
	}

	return row.ToModel(), nil
}

func (r pgRepository) UpdateSessionPeak(ctx context.Context, sessionID string, peak int32) error {
	_, err := r.db.NewUpdate().
		Model((*models.CallSessionRow)(nil)).
		Set("peak_participants = GREATEST(peak_participants, ?)", peak).
		Where("id = ?", sessionID).
		Exec(ctx)
	return err
}

func (r pgRepository) CloseSession(ctx context.Context, sessionID string, endedAt time.Time, addParticipantSeconds int64) error {
	_, err := r.db.NewUpdate().
		Model((*models.CallSessionRow)(nil)).
		Set("ended_at = ?", endedAt).
		Set("total_participant_seconds = total_participant_seconds + ?", addParticipantSeconds).
		Where("id = ? AND ended_at IS NULL", sessionID).
		Exec(ctx)
	if err != nil {
		return err
	}

	_, err = r.db.NewUpdate().
		Model((*models.CallParticipantRow)(nil)).
		Set("left_at = ?", endedAt).
		Set("duration_seconds = GREATEST(0, EXTRACT(EPOCH FROM (? - joined_at))::int)", endedAt).
		Set("disconnect_reason = CASE WHEN disconnect_reason = '' THEN 'room_closed' ELSE disconnect_reason END").
		Where("session_id = ? AND left_at IS NULL", sessionID).
		Exec(ctx)
	return err
}

func (r pgRepository) InsertParticipantJoin(ctx context.Context, participant models.CallParticipant) error {
	row := models.CallParticipantRow{
		ID:               participant.ID,
		SessionID:        participant.SessionID,
		UserID:           participant.UserID,
		JoinedAt:         participant.JoinedAt,
		ServerMuted:      participant.ServerMuted,
		ServerDeafened:   participant.ServerDeafened,
		DisconnectReason: participant.DisconnectReason,
	}
	_, err := r.db.NewInsert().Model(&row).Exec(ctx)
	return err
}

func (r pgRepository) CloseParticipantLeave(ctx context.Context, sessionID, userID string, leftAt time.Time, reason string) (int32, error) {
	var row models.CallParticipantRow
	err := r.db.NewSelect().Model(&row).
		Where("cp.session_id = ? AND cp.user_id = ? AND cp.left_at IS NULL", sessionID, userID).
		OrderExpr("cp.joined_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	dur := int32(max(0, int(leftAt.Sub(row.JoinedAt).Seconds())))
	if _, err := r.db.NewUpdate().
		Model((*models.CallParticipantRow)(nil)).
		Set("left_at = ?", leftAt).
		Set("duration_seconds = ?", dur).
		Set("disconnect_reason = ?", reason).
		Where("id = ?", row.ID).
		Exec(ctx); err != nil {
		return 0, err
	}

	if _, err := r.db.NewUpdate().
		Model((*models.CallSessionRow)(nil)).
		Set("total_participant_seconds = total_participant_seconds + ?", int64(dur)).
		Where("id = ?", sessionID).
		Exec(ctx); err != nil {
		return 0, err
	}

	return dur, nil
}

func (r pgRepository) UpdateParticipantModeration(ctx context.Context, sessionID, userID string, serverMuted, serverDeafened *bool) error {
	q := r.db.NewUpdate().Model((*models.CallParticipantRow)(nil)).
		Where("session_id = ? AND user_id = ? AND left_at IS NULL", sessionID, userID)
	if serverMuted != nil {
		q = q.Set("server_muted = ?", *serverMuted)
	}
	if serverDeafened != nil {
		q = q.Set("server_deafened = ?", *serverDeafened)
	}
	if serverMuted == nil && serverDeafened == nil {
		return nil
	}
	_, err := q.Exec(ctx)
	return err
}

func (r pgRepository) CreateRecording(ctx context.Context, rec models.CallRecording) (models.CallRecording, error) {
	row := models.CallRecordingRow{
		ID:        rec.ID,
		PublicID:  rec.PublicID,
		SessionID: rec.SessionID,
		ChannelID: rec.ChannelID,
		EgressID:  rec.EgressID,
		Status:    rec.Status,
		FileKey:   rec.FileKey,
		StartedBy: rec.StartedBy,
		StartedAt: rec.StartedAt,
	}
	if _, err := r.db.NewInsert().Model(&row).Exec(ctx); err != nil {
		return models.CallRecording{}, err
	}
	return row.ToModel(), nil
}

func (r pgRepository) FindRecordingByEgressID(ctx context.Context, egressID string) (models.CallRecording, error) {
	var row models.CallRecordingRow
	err := r.db.NewSelect().Model(&row).Where("cr.egress_id = ?", egressID).Limit(1).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CallRecording{}, ErrNotFound
		}
		return models.CallRecording{}, err
	}
	return row.ToModel(), nil
}

func (r pgRepository) FindActiveRecordingByChannel(ctx context.Context, channelID string) (models.CallRecording, error) {
	var row models.CallRecordingRow
	err := r.db.NewSelect().Model(&row).
		Where("cr.channel_id = ? AND cr.status IN ('starting', 'active', 'stopping')", channelID).
		OrderExpr("cr.started_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CallRecording{}, ErrNotFound
		}
		return models.CallRecording{}, err
	}
	return row.ToModel(), nil
}

func (r pgRepository) UpdateRecordingByEgressID(
	ctx context.Context,
	egressID, status, fileKey string,
	fileSizeBytes int64,
	durationSeconds int32,
	stoppedBy *string,
	endedAt *time.Time,
) (models.CallRecording, error) {
	var row models.CallRecordingRow
	q := r.db.NewUpdate().Model(&row).
		Set("status = ?", status).
		Where("egress_id = ?", egressID).
		Returning("*")
	if fileKey != "" {
		q = q.Set("file_key = ?", fileKey)
	}
	if fileSizeBytes > 0 {
		q = q.Set("file_size_bytes = ?", fileSizeBytes)
	}
	if durationSeconds > 0 {
		q = q.Set("duration_seconds = ?", durationSeconds)
	}
	if stoppedBy != nil {
		q = q.Set("stopped_by = ?", *stoppedBy)
	}
	if endedAt != nil {
		q = q.Set("ended_at = ?", *endedAt)
	}
	if _, err := q.Exec(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CallRecording{}, ErrNotFound
		}
		return models.CallRecording{}, err
	}
	return row.ToModel(), nil
}

func (r pgRepository) ListRecordingsByChannel(ctx context.Context, channelID string, limit int) ([]models.CallRecording, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var rows []models.CallRecordingRow
	if err := r.db.NewSelect().Model(&rows).
		Where("cr.channel_id = ?", channelID).
		OrderExpr("cr.started_at DESC").
		Limit(int64(limit)).
		Scan(ctx); err != nil {
		return nil, err
	}
	out := make([]models.CallRecording, len(rows))
	for i, row := range rows {
		out[i] = row.ToModel()
	}
	return out, nil
}

func (r pgRepository) SumParticipantSecondsBetween(ctx context.Context, from, to time.Time) (int64, error) {
	var total int64
	err := r.db.NewSelect().
		TableExpr("call_sessions AS cs").
		ColumnExpr("COALESCE(SUM(cs.total_participant_seconds), 0)").
		Where("cs.started_at >= ? AND cs.started_at < ?", from, to).
		Scan(ctx, &total)
	return total, err
}
