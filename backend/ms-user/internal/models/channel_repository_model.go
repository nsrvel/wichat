package models

import (
	"time"

	"github.com/uptrace/bun"
)

// ChannelRow maps the channels table.
type ChannelRow struct {
	bun.BaseModel `bun:"table:channels"`

	ID             string    `bun:"id,pk,type:uuid"`
	PublicID       string    `bun:"public_id"`
	GroupID        string    `bun:"group_id,type:uuid"`
	Type           string    `bun:"type"`
	IsPrivate      bool      `bun:"is_private"`
	Name           string    `bun:"name"`
	ImageURL       string    `bun:"image_url"`
	SidebarSection string    `bun:"sidebar_section"`
	CreatedBy      string    `bun:"created_by,type:uuid"`
	CreatedAt      time.Time `bun:"created_at"`
	UpdatedAt      time.Time `bun:"updated_at"`
}

// ChannelRowFrom maps a domain channel to a row.
func ChannelRowFrom(c Channel) ChannelRow {
	return ChannelRow{
		ID: c.ID, PublicID: c.PublicID, GroupID: c.GroupID, Type: c.Type, IsPrivate: c.IsPrivate,
		Name: c.Name, ImageURL: c.ImageURL, SidebarSection: c.SidebarSection,
		CreatedBy: c.CreatedBy, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

// ToModel maps a row to a domain channel.
func (r ChannelRow) ToModel() Channel {
	return Channel{
		ID: r.ID, PublicID: r.PublicID, GroupID: r.GroupID, Type: r.Type, IsPrivate: r.IsPrivate,
		Name: r.Name, ImageURL: r.ImageURL, SidebarSection: r.SidebarSection,
		CreatedBy: r.CreatedBy, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}
