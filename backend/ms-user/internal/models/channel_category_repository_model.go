package models

import (
	"time"

	"github.com/uptrace/bun"
)

// ChannelCategoryRow maps the channel_categories table.
type ChannelCategoryRow struct {
	bun.BaseModel `bun:"table:channel_categories"`

	ID        string    `bun:"id,pk,type:uuid"`
	PublicID  string    `bun:"public_id"`
	GroupID   string    `bun:"group_id,type:uuid"`
	Name      string    `bun:"name"`
	IsPrivate bool      `bun:"is_private"`
	Position  int32     `bun:"position"`
	CreatedBy string    `bun:"created_by,type:uuid"`
	CreatedAt time.Time `bun:"created_at"`
	UpdatedAt time.Time `bun:"updated_at"`
}

// ChannelCategoryRowFrom maps a domain category to a row.
func ChannelCategoryRowFrom(c ChannelCategory) ChannelCategoryRow {
	return ChannelCategoryRow{
		ID: c.ID, PublicID: c.PublicID, GroupID: c.GroupID, Name: c.Name,
		IsPrivate: c.IsPrivate, Position: c.Position,
		CreatedBy: c.CreatedBy, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

// ToModel maps a row to a domain category.
func (r ChannelCategoryRow) ToModel() ChannelCategory {
	return ChannelCategory{
		ID: r.ID, PublicID: r.PublicID, GroupID: r.GroupID, Name: r.Name,
		IsPrivate: r.IsPrivate, Position: r.Position,
		CreatedBy: r.CreatedBy, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

// ChannelCategoryPermissionOverrideRow maps channel_category_permission_overrides.
type ChannelCategoryPermissionOverrideRow struct {
	bun.BaseModel `bun:"table:channel_category_permission_overrides"`

	ID               string    `bun:"id,pk,type:uuid"`
	CategoryID       string    `bun:"category_id,type:uuid"`
	TargetType       string    `bun:"target_type"`
	TargetID         string    `bun:"target_id,type:uuid"`
	AllowPermissions []string  `bun:"allow_permissions,array"`
	DenyPermissions  []string  `bun:"deny_permissions,array"`
	CreatedAt        time.Time `bun:"created_at"`
	UpdatedAt        time.Time `bun:"updated_at"`
}
