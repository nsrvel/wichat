package repository

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

// PgRepository persists and queries universal user activity logs.
type PgRepository interface {
	Insert(ctx context.Context, log models.UserActivityLog) error
	ListByUser(
		ctx context.Context,
		userID, category, action string,
		limit int,
		beforeCreatedAt *time.Time,
		beforeID string,
	) ([]models.UserActivityLog, error)
}

type pgRepository struct {
	db *bun.DB
}

// NewPgRepository returns a Postgres-backed user activity log repository.
func NewPgRepository(db *bun.DB) PgRepository {
	return pgRepository{db: db}
}

func (r pgRepository) Insert(ctx context.Context, entry models.UserActivityLog) error {
	changesJSON, err := json.Marshal(nonNilChanges(entry.Changes))
	if err != nil {
		return err
	}
	metadataJSON, err := json.Marshal(nonNilMetadata(entry.Metadata))
	if err != nil {
		return err
	}

	var actorID *string
	if id := strings.TrimSpace(entry.ActorID); id != "" {
		actorID = &id
	}

	row := models.UserActivityLogRow{
		ID:         entry.ID,
		UserID:     entry.UserID,
		ActorID:    actorID,
		Category:   entry.Category,
		Action:     entry.Action,
		TargetType: entry.TargetType,
		TargetID:   entry.TargetID,
		Summary:    entry.Summary,
		Changes:    changesJSON,
		Metadata:   metadataJSON,
		CreatedAt:  entry.CreatedAt,
	}

	_, err = r.db.NewInsert().Model(&row).Exec(ctx)
	return err
}

func (r pgRepository) ListByUser(
	ctx context.Context,
	userID, category, action string,
	limit int,
	beforeCreatedAt *time.Time,
	beforeID string,
) ([]models.UserActivityLog, error) {
	if limit <= 0 || limit > constants.MaxActivityLogLimit {
		limit = constants.DefaultActivityLogLimit
	}

	var rows []models.UserActivityLogRow
	q := r.db.NewSelect().Model(&rows).Where("user_id = ?", userID)
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if beforeCreatedAt != nil && beforeID != "" {
		q = q.Where("(created_at, id) < (?, ?)", *beforeCreatedAt, beforeID)
	}

	err := q.OrderExpr("created_at DESC, id DESC").Limit(int64(limit)).Scan(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]models.UserActivityLog, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ToModel())
	}
	return out, nil
}

func nonNilChanges(in map[string]models.ActivityFieldChange) map[string]models.ActivityFieldChange {
	if in == nil {
		return map[string]models.ActivityFieldChange{}
	}
	return in
}

func nonNilMetadata(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	return in
}
