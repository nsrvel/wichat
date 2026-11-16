package models

import (
	"time"

	"github.com/uptrace/bun"
)

// ChannelRow maps the channels table.
type ChannelRow struct {
	bun.BaseModel `bun:"table:channels"`

	ID                    string    `bun:"id,pk,type:uuid"`
	PublicID              string    `bun:"public_id"`
	GroupID               string    `bun:"group_id,type:uuid"`
	CategoryID            *string   `bun:"category_id,type:uuid"`
	CategoryPublicID      string    `bun:"category_public_id,scanonly"`
	Type                  string    `bun:"type"`
	IsPrivate             bool      `bun:"is_private"`
	Name                  string    `bun:"name"`
	ImageURL              string    `bun:"image_url"`
	SidebarSection        string    `bun:"sidebar_section"`
	Position              int32     `bun:"position"`
	SlowmodeSeconds       int32     `bun:"slowmode_seconds"`
	ThreadSlowmodeSeconds int32     `bun:"thread_slowmode_seconds"`
	ContentVisibility     string    `bun:"content_visibility"`
	ForumDefaultLayout    string    `bun:"forum_default_layout"`
	ForumSortOrder        string    `bun:"forum_sort_order"`
	ForumTagMatching      string    `bun:"forum_tag_matching"`
	BitrateKbps           int32     `bun:"bitrate_kbps"`
	VideoQuality          string    `bun:"video_quality"`
	UserLimit             int32     `bun:"user_limit"`
	RTCRegion             string    `bun:"rtc_region"`
	CreatedBy             string    `bun:"created_by,type:uuid"`
	CreatedAt             time.Time `bun:"created_at"`
	UpdatedAt             time.Time `bun:"updated_at"`
}

// ChannelRowFrom maps a domain channel to a row.
func ChannelRowFrom(c Channel) ChannelRow {
	var categoryID *string
	if c.CategoryID != "" {
		categoryID = &c.CategoryID
	}
	contentVis := c.ContentVisibility
	if contentVis == "" {
		contentVis = "default"
	}
	forumLayout := c.ForumDefaultLayout
	if forumLayout == "" {
		forumLayout = "list"
	}
	forumSort := c.ForumSortOrder
	if forumSort == "" {
		forumSort = "recent_activity"
	}
	forumTagMatch := c.ForumTagMatching
	if forumTagMatch == "" {
		forumTagMatch = "match_some"
	}
	bitrate := c.BitrateKbps
	if bitrate == 0 {
		bitrate = 64
	}
	videoQuality := c.VideoQuality
	if videoQuality == "" {
		videoQuality = "auto"
	}
	rtcRegion := c.RTCRegion
	if rtcRegion == "" {
		rtcRegion = "automatic"
	}

	return ChannelRow{
		ID: c.ID, PublicID: c.PublicID, GroupID: c.GroupID, CategoryID: categoryID,
		Type: c.Type, IsPrivate: c.IsPrivate,
		Name: c.Name, ImageURL: c.ImageURL, SidebarSection: c.SidebarSection, Position: c.Position,
		SlowmodeSeconds: c.SlowmodeSeconds, ThreadSlowmodeSeconds: c.ThreadSlowmodeSeconds,
		ContentVisibility: contentVis, ForumDefaultLayout: forumLayout,
		ForumSortOrder: forumSort, ForumTagMatching: forumTagMatch,
		BitrateKbps:  bitrate,
		VideoQuality: videoQuality, UserLimit: c.UserLimit, RTCRegion: rtcRegion,
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
		SlowmodeSeconds: r.SlowmodeSeconds, ThreadSlowmodeSeconds: r.ThreadSlowmodeSeconds,
		ContentVisibility: r.ContentVisibility, ForumDefaultLayout: r.ForumDefaultLayout,
		ForumSortOrder: r.ForumSortOrder, ForumTagMatching: r.ForumTagMatching,
		BitrateKbps:  r.BitrateKbps,
		VideoQuality: r.VideoQuality, UserLimit: r.UserLimit, RTCRegion: r.RTCRegion,
		CreatedBy: r.CreatedBy, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

// ChannelPermissionOverrideRow maps channel_permission_overrides.
type ChannelPermissionOverrideRow struct {
	bun.BaseModel `bun:"table:channel_permission_overrides"`

	ID               string    `bun:"id,pk,type:uuid"`
	ChannelID        string    `bun:"channel_id,type:uuid"`
	TargetType       string    `bun:"target_type"`
	TargetID         string    `bun:"target_id,type:uuid"`
	AllowPermissions []string  `bun:"allow_permissions,array"`
	DenyPermissions  []string  `bun:"deny_permissions,array"`
	CreatedAt        time.Time `bun:"created_at"`
	UpdatedAt        time.Time `bun:"updated_at"`
}
