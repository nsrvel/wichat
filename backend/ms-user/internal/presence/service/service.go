package service

import (
	"context"
	"strings"
	"time"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	presencerepo "github.com/wichat/wichat/backend/ms-user/internal/presence/repository"
	"github.com/wichat/wichat/backend/ms-user/pkg/access"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

const defaultPresenceTTL = 60 * time.Second

// Reader provides batch presence snapshots for member listings.
type Reader interface {
	BatchGetEffectivePresence(ctx context.Context, userIDs []string) (map[string]models.UserPresenceSnapshot, error)
}

// Service manages WebSocket heartbeats, manual presence overrides, and effective status resolution.
type Service interface {
	Reader
	HeartbeatPresence(ctx context.Context, accessToken, connID, activity string) (models.PresenceTransition, error)
	DisconnectPresence(ctx context.Context, userID, connID string) (models.PresenceTransition, error)
	SetManualPresence(ctx context.Context, accessToken, status string, ttlSeconds int64) (models.SetManualPresenceResult, error)
	GetMyPresence(ctx context.Context, accessToken string) (models.MyPresenceView, error)
}

type service struct {
	repo     presencerepo.RedisRepository
	groups   grouprepo.PgRepository
	sessions authrepo.SessionRepository
	secret   []byte
	log      logger.Logger
	now      func() time.Time
}

// NewService constructs the presence service.
func NewService(
	repo presencerepo.RedisRepository,
	groups grouprepo.PgRepository,
	sessions authrepo.SessionRepository,
	secret []byte,
	log logger.Logger,
) Service {
	copied := make([]byte, len(secret))
	copy(copied, secret)

	return service{
		repo:     repo,
		groups:   groups,
		sessions: sessions,
		secret:   copied,
		log:      log,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func normalizeActivity(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), models.PresenceActivityIdle) {
		return models.PresenceActivityIdle
	}
	return models.PresenceActivityActive
}

func isValidManualStatus(status string) bool {
	switch status {
	case models.PresenceStatusOnline,
		models.PresenceStatusIdle,
		models.PresenceStatusDND,
		models.PresenceStatusInvisible:
		return true
	default:
		return false
	}
}

func resolveEffectiveStatus(manualStatus string, activities []string) string {
	if len(activities) == 0 {
		return models.PresenceStatusOffline
	}
	switch manualStatus {
	case models.PresenceStatusInvisible:
		return models.PresenceStatusOffline
	case models.PresenceStatusDND:
		return models.PresenceStatusDND
	case models.PresenceStatusIdle:
		return models.PresenceStatusIdle
	default:
		for _, a := range activities {
			if a == models.PresenceActivityActive {
				return models.PresenceStatusOnline
			}
		}
		return models.PresenceStatusIdle
	}
}

func (s service) resolveUserEffectivePresence(ctx context.Context, userID string) (models.UserPresenceSnapshot, error) {
	effective, err := s.repo.GetEffectiveStatus(ctx, userID)
	if err != nil {
		return models.UserPresenceSnapshot{}, err
	}
	snap := models.UserPresenceSnapshot{
		IsOnline:       effective != models.PresenceStatusOffline,
		PresenceStatus: effective,
	}
	if snap.IsOnline {
		return snap, nil
	}

	activities, actErr := s.repo.ListActiveConnActivities(ctx, userID)
	if actErr != nil || len(activities) == 0 {
		return snap, nil
	}

	manual, manualErr := s.repo.GetManualStatus(ctx, userID)
	if manualErr != nil {
		return models.UserPresenceSnapshot{}, manualErr
	}

	effective = resolveEffectiveStatus(manual.Status, activities)
	snap = models.UserPresenceSnapshot{
		IsOnline:       effective != models.PresenceStatusOffline,
		PresenceStatus: effective,
	}
	if snap.IsOnline {
		_ = s.repo.SetEffectiveStatus(ctx, userID, effective, defaultPresenceTTL)
	}
	return snap, nil
}

func (s service) userGroupPublicIDs(ctx context.Context, userID string) []string {
	if s.groups == nil {
		return nil
	}
	groups, err := s.groups.ListGroupsForUser(ctx, userID)
	if err != nil {
		return nil
	}
	ids := make([]string, 0, len(groups))
	for _, g := range groups {
		if g.PublicID != "" {
			ids = append(ids, g.PublicID)
		}
	}
	return ids
}

func (s service) recomputeTransition(ctx context.Context, userID, prevStatus string) (models.PresenceTransition, error) {
	manual, err := s.repo.GetManualStatus(ctx, userID)
	if err != nil {
		return models.PresenceTransition{}, err
	}
	activities, err := s.repo.ListActiveConnActivities(ctx, userID)
	if err != nil {
		return models.PresenceTransition{}, err
	}

	effective := resolveEffectiveStatus(manual.Status, activities)
	if err := s.repo.SetEffectiveStatus(ctx, userID, effective, defaultPresenceTTL); err != nil {
		return models.PresenceTransition{}, err
	}

	var groupIDs []string
	if prevStatus != effective {
		groupIDs = s.userGroupPublicIDs(ctx, userID)
	}

	return models.PresenceTransition{
		UserID:          userID,
		PreviousStatus:  prevStatus,
		EffectiveStatus: effective,
		IsOnline:        effective != models.PresenceStatusOffline,
		GroupIDs:        groupIDs,
	}, nil
}

func (s service) HeartbeatPresence(ctx context.Context, accessToken, connID, activity string) (models.PresenceTransition, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return models.PresenceTransition{}, err
	}
	connID = strings.TrimSpace(connID)
	if connID == "" {
		return models.PresenceTransition{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "conn_id"})
	}

	prevStatus, err := s.repo.GetEffectiveStatus(ctx, userID)
	if err != nil {
		return models.PresenceTransition{}, err
	}

	if err := s.repo.SetConnHeartbeat(ctx, userID, connID, normalizeActivity(activity), defaultPresenceTTL); err != nil {
		return models.PresenceTransition{}, err
	}

	return s.recomputeTransition(ctx, userID, prevStatus)
}

func (s service) DisconnectPresence(ctx context.Context, userID, connID string) (models.PresenceTransition, error) {
	userID = strings.TrimSpace(userID)
	connID = strings.TrimSpace(connID)
	if userID == "" || connID == "" {
		return models.PresenceTransition{}, nil
	}

	prevStatus, err := s.repo.GetEffectiveStatus(ctx, userID)
	if err != nil {
		return models.PresenceTransition{}, err
	}

	if err := s.repo.DeleteConn(ctx, userID, connID); err != nil {
		return models.PresenceTransition{}, err
	}

	return s.recomputeTransition(ctx, userID, prevStatus)
}

func (s service) SetManualPresence(ctx context.Context, accessToken, status string, ttlSeconds int64) (models.SetManualPresenceResult, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return models.SetManualPresenceResult{}, err
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if !isValidManualStatus(status) {
		return models.SetManualPresenceResult{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "status"})
	}
	if ttlSeconds < 0 {
		return models.SetManualPresenceResult{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "ttl_seconds"})
	}

	prevStatus, err := s.repo.GetEffectiveStatus(ctx, userID)
	if err != nil {
		return models.SetManualPresenceResult{}, err
	}

	var ttl time.Duration
	if ttlSeconds > 0 {
		ttl = time.Duration(ttlSeconds) * time.Second
	}
	if err := s.repo.SetManualStatus(ctx, userID, status, ttl); err != nil {
		return models.SetManualPresenceResult{}, err
	}

	tr, err := s.recomputeTransition(ctx, userID, prevStatus)
	if err != nil {
		return models.SetManualPresenceResult{}, err
	}

	manual, err := s.repo.GetManualStatus(ctx, userID)
	if err != nil {
		return models.SetManualPresenceResult{}, err
	}

	return models.SetManualPresenceResult{
		UserID:          userID,
		ManualStatus:    manual.Status,
		EffectiveStatus: tr.EffectiveStatus,
		IsOnline:        tr.IsOnline,
		ExpiresAt:       manual.ExpiresAt,
		GroupIDs:        tr.GroupIDs,
	}, nil
}

func degradedMyPresenceView(manual models.ManualPresenceState) models.MyPresenceView {
	manualStatus := models.PresenceStatusOnline
	if manual.Status != "" {
		manualStatus = manual.Status
	}
	return models.MyPresenceView{
		ManualStatus:    manualStatus,
		EffectiveStatus: models.PresenceStatusOffline,
		IsOnline:        false,
		ExpiresAt:       manual.ExpiresAt,
	}
}

func offlinePresenceSnapshots(userIDs []string) map[string]models.UserPresenceSnapshot {
	out := make(map[string]models.UserPresenceSnapshot, len(userIDs))
	for _, id := range userIDs {
		out[id] = models.UserPresenceSnapshot{
			IsOnline:       false,
			PresenceStatus: models.PresenceStatusOffline,
		}
	}
	return out
}

func (s service) GetMyPresence(ctx context.Context, accessToken string) (models.MyPresenceView, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return models.MyPresenceView{}, err
	}

	manual, err := s.repo.GetManualStatus(ctx, userID)
	if err != nil {
		if s.log != nil {
			s.log.Warn("presence manual read failed, degrading", "err", err, "user_id", userID)
		}
		return degradedMyPresenceView(models.ManualPresenceState{Status: models.PresenceStatusOnline}), nil
	}
	snap, err := s.resolveUserEffectivePresence(ctx, userID)
	if err != nil {
		if s.log != nil {
			s.log.Warn("presence effective resolve failed, degrading", "err", err, "user_id", userID)
		}
		return degradedMyPresenceView(manual), nil
	}

	return models.MyPresenceView{
		ManualStatus:    manual.Status,
		EffectiveStatus: snap.PresenceStatus,
		IsOnline:        snap.IsOnline,
		ExpiresAt:       manual.ExpiresAt,
	}, nil
}

func (s service) BatchGetEffectivePresence(ctx context.Context, userIDs []string) (map[string]models.UserPresenceSnapshot, error) {
	out, err := s.repo.BatchGetEffectivePresence(ctx, userIDs)
	if err != nil {
		if s.log != nil {
			s.log.Warn("presence batch read failed, degrading to offline", "err", err)
		}
		return offlinePresenceSnapshots(userIDs), nil
	}
	for _, id := range userIDs {
		snap, ok := out[id]
		if !ok || snap.IsOnline {
			continue
		}
		resolved, resolveErr := s.resolveUserEffectivePresence(ctx, id)
		if resolveErr != nil {
			continue
		}
		out[id] = resolved
	}
	return out, nil
}
