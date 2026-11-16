package models

import (
	"time"

	"github.com/uptrace/bun"
)

// ChannelRow maps the channels table.
type ChannelRow struct {
	bun.BaseModel `bun:"table:channels"`

	ID               string    `bun:"id,pk,type:uuid"`
	PublicID         string    `bun:"public_id"`
	GroupID          string    `bun:"group_id,type:uuid"`
	CategoryID       *string   `bun:"category_id,type:uuid"`
	CategoryPublicID string    `bun:"category_public_id,scanonly"`
	Type             string    `bun:"type"`
	IsPrivate        bool      `bun:"is_private"`
	Name             string    `bun:"name"`
	ImageURL         string    `bun:"image_url"`
	SidebarSection   string    `bun:"sidebar_section"`
	Position         int32     `bun:"position"`
	CreatedBy        string    `bun:"created_by,type:uuid"`
	CreatedAt        time.Time `bun:"created_at"`
	UpdatedAt        time.Time `bun:"updated_at"`
}

// ChannelRowFrom maps a domain channel to a row.
func ChannelRowFrom(c Channel) ChannelRow {
	var categoryID *string
	if c.CategoryID != "" {
		categoryID = &c.CategoryID
	}

	return ChannelRow{
		ID: c.ID, PublicID: c.PublicID, GroupID: c.GroupID, CategoryID: categoryID,
		Type: c.Type, IsPrivate: c.IsPrivate,
		Name: c.Name, ImageURL: c.ImageURL, SidebarSection: c.SidebarSection, Position: c.Position,
		CreatedBy: c.CreatedBy, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

// ToModel maps a row to a domain channel.
func (r ChannelRow) ToModel() Channel {
	var categoryID string
	if r.CategoryID != nil {
		categoryID = *r.CategoryID
	}

	return Channel{
		ID: r.ID, PublicID: r.PublicID, GroupID: r.GroupID,
		CategoryID: categoryID, CategoryPublicID: r.CategoryPublicID,
		Type: r.Type, IsPrivate: r.IsPrivate,
		Name: r.Name, ImageURL: r.ImageURL, SidebarSection: r.SidebarSection, Position: r.Position,
		CreatedBy: r.CreatedBy, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

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
