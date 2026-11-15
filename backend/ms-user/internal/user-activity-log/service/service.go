package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/internal/user-activity-log/repository"
	"github.com/wichat/wichat/backend/ms-user/pkg/access"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// Recorder records universal user audit events without failing the primary user action.
type Recorder interface {
	Record(ctx context.Context, entry models.UserActivityLogEntry)
}

// Service records and queries universal user activity logs.
type Service interface {
	Recorder
	ListMyActivityLogs(ctx context.Context, accessToken, category, action string, limit int, cursor string) (models.UserActivityLogPage, error)
}

type service struct {
	repo     repository.PgRepository
	sessions authrepo.SessionRepository
	secret   []byte
	log      logger.Logger
	now      func() time.Time
}

// NewService constructs the user-activity-log service.
func NewService(
	repo repository.PgRepository,
	sessions authrepo.SessionRepository,
	secret []byte,
	log logger.Logger,
) Service {
	return service{
		repo:     repo,
		sessions: sessions,
		secret:   secret,
		log:      log,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (s service) Record(ctx context.Context, entry models.UserActivityLogEntry) {
	if s.repo == nil || strings.TrimSpace(entry.UserID) == "" || strings.TrimSpace(entry.Action) == "" {
		return
	}

	ts := entry.CreatedAt.UTC()
	if ts.IsZero() {
		ts = s.now()
	}

	actorID := strings.TrimSpace(entry.ActorID)
	if actorID == "" {
		actorID = entry.UserID
	}

	record := models.UserActivityLog{
		ID:         uuid.NewString(),
		UserID:     entry.UserID,
		ActorID:    actorID,
		Category:   strings.TrimSpace(entry.Category),
		Action:     strings.TrimSpace(entry.Action),
		TargetType: strings.TrimSpace(entry.TargetType),
		TargetID:   strings.TrimSpace(entry.TargetID),
		Summary:    strings.TrimSpace(entry.Summary),
		Changes:    entry.Changes,
		Metadata:   entry.Metadata,
		CreatedAt:  ts,
	}

	if err := s.repo.Insert(ctx, record); err != nil && s.log != nil {
		s.log.Warn("record user activity log failed", "user_id", entry.UserID, "action", entry.Action, "error", err)
	}
}

func (s service) ListMyActivityLogs(
	ctx context.Context,
	accessToken,
	category,
	action string,
	limit int,
	cursor string,
) (models.UserActivityLogPage, error) {
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return models.UserActivityLogPage{}, err
	}

	if limit <= 0 {
		limit = constants.DefaultActivityLogLimit
	}
	if limit > constants.MaxActivityLogLimit {
		limit = constants.MaxActivityLogLimit
	}

	var beforeCreatedAt *time.Time
	var beforeID string
	if c := strings.TrimSpace(cursor); c != "" {
		ts, id, err := decodeCursor(c)
		if err != nil {
			return models.UserActivityLogPage{}, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "cursor"})
		}
		beforeCreatedAt = &ts
		beforeID = id
	}

	rows, err := s.repo.ListByUser(
		ctx,
		userID,
		strings.TrimSpace(category),
		strings.TrimSpace(action),
		limit+1,
		beforeCreatedAt,
		beforeID,
	)
	if err != nil {
		return models.UserActivityLogPage{}, err
	}

	nextCursor := ""
	if len(rows) > limit {
		last := rows[limit-1]
		nextCursor = encodeCursor(last.CreatedAt, last.ID)
		rows = rows[:limit]
	}

	return models.UserActivityLogPage{
		Logs:       rows,
		NextCursor: nextCursor,
	}, nil
}

func encodeCursor(ts time.Time, id string) string {
	raw := fmt.Sprintf("%s|%s", ts.UTC().Format(time.RFC3339Nano), id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(cursor string) (time.Time, string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", err
	}
	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 || parts[1] == "" {
		return time.Time{}, "", fmt.Errorf("invalid cursor")
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", err
	}
	return ts.UTC(), parts[1], nil
}
