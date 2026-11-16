package constants

const (
	ChannelKindBroadcast = "broadcast"
	ChannelKindTextOnly  = "text_only"
	ChannelKindVoiceOnly = "voice_only"
	ChannelKindForum     = "forum"
	ChannelKindDM        = "dm"

	DefaultChannelName             = "General"
	DefaultChannelLookupKey        = "general"
	DefaultAnnouncementChannelName = "Announcements"
	DefaultTextCategoryName        = "Text Channels"
	DefaultVoiceCategoryName       = "Voice Channels"

	// SidebarSection* drives client sidebar grouping (see channels.sidebar_section).
	SidebarSectionChannels  = "channels"
	SidebarSectionFavorites = "favorites"
)
