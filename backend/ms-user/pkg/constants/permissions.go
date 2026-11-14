package constants

// Group role slugs.
const (
	RoleSlugEveryone = "everyone"
	RoleSlugMember   = "member"
)

// Permissions in WiChat RBAC.
const (
	PermAdministrator        = "administrator"
	PermViewChannels         = "view_channels"
	PermManageChannels       = "manage_channels"
	PermManageRoles          = "manage_roles"
	PermManageGroup          = "manage_group"
	PermCreateExpressions    = "create_expressions"
	PermManageExpressions    = "manage_expressions"
	PermViewAuditLog         = "view_audit_log"
	PermViewGroupInsights    = "view_group_insights"
	PermManageWebhooks       = "manage_webhooks"
	PermCreateInvite         = "create_invite"
	PermKickMembers          = "kick_members"
	PermBanMembers           = "ban_members"
	PermTimeoutMembers       = "timeout_members"
	PermSendMessages         = "send_messages"
	PermSendMessagesThreads  = "send_messages_in_threads"
	PermCreatePublicThreads  = "create_public_threads"
	PermCreatePrivateThreads = "create_private_threads"
	PermEmbedLinks           = "embed_links"
	PermAttachFiles          = "attach_files"
	PermAddReactions         = "add_reactions"
	PermMentionEveryone      = "mention_everyone"
	PermManageMessages       = "manage_messages"
	PermPinMessages          = "pin_messages"
	PermBypassSlowmode       = "bypass_slowmode"
	PermManageThreads        = "manage_threads"
	PermReadMessageHistory   = "read_message_history"
	PermSendVoiceMessages    = "send_voice_messages"
	PermCreatePolls          = "create_polls"
	PermConnect              = "connect"
	PermSpeak                = "speak"
	PermVideo                = "video"
	PermUseVoiceActivity     = "use_voice_activity"
	PermUseSoundboard        = "use_soundboard"
	PermPrioritySpeaker      = "priority_speaker"
	PermMuteMembers          = "mute_members"
	PermDeafenMembers        = "deafen_members"
	PermMoveMembers          = "move_members"
	PermManagePlugins        = "manage_plugins"
	PermUsePlugins           = "use_plugins"
	PermCreateEvents         = "create_events"
	PermManageEvents         = "manage_events"
)

// AllKnownPermissions contains all valid permissions recognized by the platform.
var AllKnownPermissions = []string{
	PermAdministrator,
	PermViewChannels,
	PermManageChannels,
	PermManageRoles,
	PermManageGroup,
	PermCreateExpressions,
	PermManageExpressions,
	PermViewAuditLog,
	PermViewGroupInsights,
	PermManageWebhooks,
	PermCreateInvite,
	PermKickMembers,
	PermBanMembers,
	PermTimeoutMembers,
	PermSendMessages,
	PermSendMessagesThreads,
	PermCreatePublicThreads,
	PermCreatePrivateThreads,
	PermEmbedLinks,
	PermAttachFiles,
	PermAddReactions,
	PermMentionEveryone,
	PermManageMessages,
	PermPinMessages,
	PermBypassSlowmode,
	PermManageThreads,
	PermReadMessageHistory,
	PermSendVoiceMessages,
	PermCreatePolls,
	PermConnect,
	PermSpeak,
	PermVideo,
	PermUseVoiceActivity,
	PermUseSoundboard,
	PermPrioritySpeaker,
	PermMuteMembers,
	PermDeafenMembers,
	PermMoveMembers,
	PermManagePlugins,
	PermUsePlugins,
	PermCreateEvents,
	PermManageEvents,
}

// DefaultEveryonePermissions defines the out-of-the-box permissions granted to @everyone.
var DefaultEveryonePermissions = []string{
	PermViewChannels,
	PermSendMessages,
	PermReadMessageHistory,
	PermCreateInvite,
	PermAttachFiles,
	PermAddReactions,
	PermConnect,
	PermSpeak,
	PermUseVoiceActivity,
}
