package models

import (
	"time"

	"github.com/uptrace/bun"
)

// ChannelRow maps the channels table.
type ChannelRow struct {
	bun.BaseModel `bun:"table:channels"`

	ID        string    `bun:"id,pk,type:uuid"`
	PublicID  string    `bun:"public_id"`
	GroupID   string    `bun:"group_id,type:uuid"`
	Type      string    `bun:"type"`
	Name      string    `bun:"name"`
	CreatedBy string    `bun:"created_by,type:uuid"`
	CreatedAt time.Time `bun:"created_at"`
	UpdatedAt time.Time `bun:"updated_at"`
}

// ChannelRowFrom maps a domain channel to a row.
func ChannelRowFrom(c Channel) ChannelRow {
	return ChannelRow{
		ID: c.ID, PublicID: c.PublicID, GroupID: c.GroupID, Type: c.Type, Name: c.Name,
		CreatedBy: c.CreatedBy, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

// ToModel maps a row to a domain channel.
func (r ChannelRow) ToModel() Channel {
	return Channel{
		ID: r.ID, PublicID: r.PublicID, GroupID: r.GroupID, Type: r.Type, Name: r.Name,
		CreatedBy: r.CreatedBy, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}
