package models

import "time"

// Channel is workspace channel metadata (ms-user).
type Channel struct {
	ID                    string
	PublicID              string
	GroupID               string
	CategoryID            string
	CategoryPublicID      string
	Type                  string
	IsPrivate             bool
	Name                  string
	ImageURL              string
	SidebarSection        string
	Position              int32
	SlowmodeSeconds       int32
	ThreadSlowmodeSeconds int32
	ContentVisibility     string
	ForumDefaultLayout    string
	ForumSortOrder        string
	ForumTagMatching      string
	BitrateKbps           int32
	VideoQuality          string
	UserLimit             int32
	RTCRegion             string
	CreatedBy             string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// UpdateChannelInput holds optional fields for updating a channel.
type UpdateChannelInput struct {
	Name                  *string
	IsPrivate             *bool
	SlowmodeSeconds       *int32
	ThreadSlowmodeSeconds *int32
	ContentVisibility     *string
	ForumDefaultLayout    *string
	ForumSortOrder        *string
	ForumTagMatching      *string
	BitrateKbps           *int32
	VideoQuality          *string
	UserLimit             *int32
	RTCRegion             *string
}

// ChannelAccess holds internal ids and effective metadata after token, membership, and channel resolution.
type ChannelAccess struct {
	ChannelID         string
	UserID            string
	Permissions       []string
	ChannelType       string
	SlowmodeSeconds   int32
	ContentVisibility string
}

// PublicChannel is returned to clients (JSON id is public_id).
type PublicChannel struct {
	ID                    string
	Name                  string
	Type                  string
	IsPrivate             bool
	ImageURL              string
	SidebarSection        string
	CategoryID            string
	Position              int32
	SlowmodeSeconds       int32
	ThreadSlowmodeSeconds int32
	ContentVisibility     string
	ForumDefaultLayout    string
	ForumSortOrder        string
	ForumTagMatching      string
	BitrateKbps           int32
	VideoQuality          string
	UserLimit             int32
	RTCRegion             string
	CallerPermissions     []string
	DmPeerUserID          string
}

// Public returns client-safe channel fields.
func (c Channel) Public() PublicChannel {
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

	return PublicChannel{
		ID: c.PublicID, Name: c.Name, Type: c.Type, IsPrivate: c.IsPrivate,
		ImageURL: c.ImageURL, SidebarSection: c.SidebarSection, CategoryID: c.CategoryPublicID,
		Position: c.Position, SlowmodeSeconds: c.SlowmodeSeconds, ThreadSlowmodeSeconds: c.ThreadSlowmodeSeconds,
		ContentVisibility: contentVis, ForumDefaultLayout: forumLayout, ForumSortOrder: forumSort,
		ForumTagMatching: forumTagMatch, BitrateKbps: bitrate,
		VideoQuality: videoQuality, UserLimit: c.UserLimit, RTCRegion: rtcRegion,
	}
}

// ChannelPermissionOverride is internal channel permission override metadata.
type ChannelPermissionOverride struct {
	ChannelID        string
	TargetType       string
	TargetID         string
	IsEveryone       bool
	AllowPermissions []string
	DenyPermissions  []string
}
