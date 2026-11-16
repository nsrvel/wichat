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

// PgRepository persists and queries group activity logs.
type PgRepository interface {
	Insert(ctx context.Context, log models.GroupActivityLog) error
	ListByGroup(
		ctx context.Context,
		filter models.GroupActivityLogFilter,
		limit int,
		beforeCreatedAt *time.Time,
		beforeID string,
	) ([]models.GroupActivityLog, error)
}

type pgRepository struct {
	db *bun.DB
}

// NewPgRepository returns a Postgres-backed group activity log repository.
func NewPgRepository(db *bun.DB) PgRepository {
	return pgRepository{db: db}
}

func (r pgRepository) Insert(ctx context.Context, entry models.GroupActivityLog) error {
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

	row := models.GroupActivityLogRow{
		ID:         entry.ID,
		GroupID:    entry.GroupID,
		ActorID:    actorID,
		Category:   entry.Category,
		Action:     entry.Action,
		TargetType: entry.TargetType,
		TargetID:   entry.TargetID,
		TargetName: entry.TargetName,
		Summary:    entry.Summary,
		Changes:    changesJSON,
		Metadata:   metadataJSON,
		CreatedAt:  entry.CreatedAt,
	}

	_, err = r.db.NewInsert().Model(&row).Exec(ctx)
	return err
}

func (r pgRepository) ListByGroup(
	ctx context.Context,
	filter models.GroupActivityLogFilter,
	limit int,
	beforeCreatedAt *time.Time,
	beforeID string,
) ([]models.GroupActivityLog, error) {
	if limit <= 0 || limit > constants.MaxActivityLogLimit {
		limit = constants.DefaultActivityLogLimit
	}

	var rows []models.GroupActivityLogReadRow
	q := r.db.NewSelect().
		Model(&rows).
		ColumnExpr("gal.id, gal.group_id, gal.actor_id, gal.category, gal.action, gal.target_type, gal.target_id, gal.target_name, gal.summary, gal.changes, gal.metadata, gal.created_at").
		ColumnExpr("COALESCE(u.username, '') AS actor_username").
		ColumnExpr("COALESCE(u.display_name, '') AS actor_display_name").
		ColumnExpr("COALESCE(u.avatar_url, '') AS actor_avatar_url").
		Join("LEFT JOIN users u ON u.id = gal.actor_id").
		Where("gal.group_id = ?", filter.GroupID)

	if filter.ActorID != "" {
		q = q.Where("gal.actor_id = ?", filter.ActorID)
	}
	if filter.Category != "" {
		q = q.Where("gal.category = ?", filter.Category)
	}
	if filter.Action != "" {
		q = q.Where("gal.action = ?", filter.Action)
	}
	if beforeCreatedAt != nil && beforeID != "" {
		q = q.Where("(gal.created_at, gal.id) < (?, ?)", *beforeCreatedAt, beforeID)
	}

	err := q.OrderExpr("gal.created_at DESC, gal.id DESC").Limit(int64(limit)).Scan(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]models.GroupActivityLog, 0, len(rows))
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
