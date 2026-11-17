package constants

const (
	ChannelKindBroadcast    = "broadcast"
	ChannelKindTextOnly     = "text_only"
	ChannelKindVoiceOnly    = "voice_only"
	ChannelKindTextAndVoice = "text_and_voice"
	ChannelKindForum        = "forum"
	ChannelKindDM           = "dm"

	DefaultChannelName             = "General"
	DefaultChannelLookupKey        = "general"
	DefaultAnnouncementChannelName = "Announcements"
	DefaultFavoritesCategoryName   = "Favorites"
	DefaultTextCategoryName        = "Text Channels"
	DefaultVoiceCategoryName       = "Voice Channels"

	// SidebarSection* drives client sidebar grouping (see channels.sidebar_section).
	SidebarSectionChannels  = "channels"
	SidebarSectionFavorites = "favorites"
)
