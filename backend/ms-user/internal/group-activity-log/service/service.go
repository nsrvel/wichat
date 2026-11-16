package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/group-activity-log/repository"
	groupmemberrepo "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository"
	grouprolerepo "github.com/wichat/wichat/backend/ms-user/internal/group-role/repository"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/access"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// Recorder records group audit events without failing the primary group action.
type Recorder interface {
	Record(ctx context.Context, entry models.GroupActivityLogEntry)
}

// Service records and queries group activity logs.
type Service interface {
	Recorder
	ListGroupActivityLogs(
		ctx context.Context,
		accessToken, groupPublicID, actorID, category, action string,
		limit int,
		cursor string,
	) (*models.GroupActivityLogPage, error)
}

type service struct {
	repo     repository.PgRepository
	groups   grouprepo.PgRepository
	members  groupmemberrepo.PgRepository
	roles    grouprolerepo.PgRepository
	sessions authrepo.SessionRepository
	secret   []byte
	log      logger.Logger
	now      func() time.Time
}

// NewService constructs the group-activity-log service.
func NewService(
	repo repository.PgRepository,
	groups grouprepo.PgRepository,
	members groupmemberrepo.PgRepository,
	roles grouprolerepo.PgRepository,
	sessions authrepo.SessionRepository,
	secret []byte,
	log logger.Logger,
) Service {
	return service{
		repo:     repo,
		groups:   groups,
		members:  members,
		roles:    roles,
		sessions: sessions,
		secret:   secret,
		log:      log,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (s service) Record(ctx context.Context, entry models.GroupActivityLogEntry) {
	if s.repo == nil || strings.TrimSpace(entry.GroupID) == "" || strings.TrimSpace(entry.Action) == "" {
		return
	}

	ts := entry.CreatedAt.UTC()
	if ts.IsZero() {
		ts = s.now()
	}

	record := models.GroupActivityLog{
		ID:         uuid.NewString(),
		GroupID:    strings.TrimSpace(entry.GroupID),
		ActorID:    strings.TrimSpace(entry.ActorID),
		Category:   strings.TrimSpace(entry.Category),
		Action:     strings.TrimSpace(entry.Action),
		TargetType: strings.TrimSpace(entry.TargetType),
		TargetID:   strings.TrimSpace(entry.TargetID),
		TargetName: strings.TrimSpace(entry.TargetName),
		Summary:    strings.TrimSpace(entry.Summary),
		Changes:    entry.Changes,
		Metadata:   entry.Metadata,
		CreatedAt:  ts,
	}

	if err := s.repo.Insert(ctx, record); err != nil && s.log != nil {
		s.log.Warn("record group activity log failed", "group_id", entry.GroupID, "action", entry.Action, "error", err)
	}
}

func (s service) ListGroupActivityLogs(
	ctx context.Context,
	accessToken, groupPublicID, actorID, category, action string,
	limit int,
	cursor string,
) (*models.GroupActivityLogPage, error) {
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.now())
	if err != nil {
		return nil, err
	}
	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return nil, err
	}
	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		return nil, err
	}
	member, err := s.members.FindGroupMember(ctx, group.ID, userID)
	if err != nil {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}
	perms, err := s.roles.GetEffectivePermissions(ctx, group.ID, userID)
	if err != nil {
		return nil, err
	}
	if !access.HasPermission(member.Role == constants.GroupRoleOwner, perms, constants.PermViewAuditLog) {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
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
			return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "cursor"})
		}
		beforeCreatedAt = &ts
		beforeID = id
	}

	rows, err := s.repo.ListByGroup(
		ctx,
		models.GroupActivityLogFilter{
			GroupID:  group.ID,
			ActorID:  strings.TrimSpace(actorID),
			Category: strings.TrimSpace(category),
			Action:   strings.TrimSpace(action),
		},
		limit+1,
		beforeCreatedAt,
		beforeID,
	)
	if err != nil {
		return nil, err
	}

	nextCursor := ""
	if len(rows) > limit {
		last := rows[limit-1]
		nextCursor = encodeCursor(last.CreatedAt, last.ID)
		rows = rows[:limit]
	}

	return &models.GroupActivityLogPage{
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
