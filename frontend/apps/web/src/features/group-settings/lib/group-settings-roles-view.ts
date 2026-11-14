/** Built-in membership roles (custom roles API deferred). */
export type GroupSettingsBuiltinRoleId = "member" | "everyone";

export type GroupSettingsRoleDetailTab = "display" | "permissions" | "members";

export type GroupSettingsRoleRow = {
	id: string;
	/** Stable key for built-ins; display name for custom roles later. */
	name: string;
	memberCount: number;
	isSystem: boolean;
	color: string;
};

export const DEFAULT_ROLE_COLOR = "#99aab5";

/** 20 preset role colors (2 rows × 10 columns) matching standard role palette. */
export const ROLE_COLOR_PRESETS = [
	"#1abc9c",
	"#2ecc71",
	"#3498db",
	"#9b59b6",
	"#e91e63",
	"#f1c40f",
	"#e67e22",
	"#e74c3c",
	"#95a5a6",
	"#607d8b",
	"#11806a",
	"#1f8b4c",
	"#206694",
	"#71368a",
	"#ad1457",
	"#c27c0e",
	"#a84300",
	"#992d22",
	"#979c9f",
	"#546e7a",
] as const;

export type RolePermissionItem = {
	id: string;
	titleKey: string;
	descriptionKey: string;
	tier?: "core" | "next";
	dangerous?: boolean;
};

export type RolePermissionGroup = {
	id: string;
	titleKey: string;
	items: RolePermissionItem[];
};

export const ROLE_PERMISSION_GROUPS: RolePermissionGroup[] = [
	{
		id: "general",
		titleKey: "group_settings.roles.detail.perm_group_general",
		items: [
			{
				id: "view_channels",
				titleKey: "group_settings.roles.detail.perm_view_channels",
				descriptionKey: "group_settings.roles.detail.perm_view_channels_desc",
				tier: "core",
			},
			{
				id: "manage_channels",
				titleKey: "group_settings.roles.detail.perm_manage_channels",
				descriptionKey: "group_settings.roles.detail.perm_manage_channels_desc",
				tier: "core",
			},
			{
				id: "manage_roles",
				titleKey: "group_settings.roles.detail.perm_manage_roles",
				descriptionKey: "group_settings.roles.detail.perm_manage_roles_desc",
				tier: "core",
			},
			{
				id: "manage_group",
				titleKey: "group_settings.roles.detail.perm_manage_group",
				descriptionKey: "group_settings.roles.detail.perm_manage_group_desc",
				tier: "core",
			},
			{
				id: "create_expressions",
				titleKey: "group_settings.roles.detail.perm_create_expressions",
				descriptionKey: "group_settings.roles.detail.perm_create_expressions_desc",
				tier: "next",
			},
			{
				id: "manage_expressions",
				titleKey: "group_settings.roles.detail.perm_manage_expressions",
				descriptionKey: "group_settings.roles.detail.perm_manage_expressions_desc",
				tier: "next",
			},
			{
				id: "view_audit_log",
				titleKey: "group_settings.roles.detail.perm_view_audit_log",
				descriptionKey: "group_settings.roles.detail.perm_view_audit_log_desc",
				tier: "next",
			},
			{
				id: "view_group_insights",
				titleKey: "group_settings.roles.detail.perm_view_group_insights",
				descriptionKey: "group_settings.roles.detail.perm_view_group_insights_desc",
				tier: "next",
			},
			{
				id: "manage_webhooks",
				titleKey: "group_settings.roles.detail.perm_manage_webhooks",
				descriptionKey: "group_settings.roles.detail.perm_manage_webhooks_desc",
				tier: "next",
			},
		],
	},
	{
		id: "membership",
		titleKey: "group_settings.roles.detail.perm_group_membership",
		items: [
			{
				id: "create_invite",
				titleKey: "group_settings.roles.detail.perm_create_invite",
				descriptionKey: "group_settings.roles.detail.perm_create_invite_desc",
				tier: "core",
			},
			{
				id: "kick_members",
				titleKey: "group_settings.roles.detail.perm_kick_members",
				descriptionKey: "group_settings.roles.detail.perm_kick_members_desc",
				tier: "core",
			},
			{
				id: "ban_members",
				titleKey: "group_settings.roles.detail.perm_ban_members",
				descriptionKey: "group_settings.roles.detail.perm_ban_members_desc",
				tier: "core",
			},
			{
				id: "timeout_members",
				titleKey: "group_settings.roles.detail.perm_timeout_members",
				descriptionKey: "group_settings.roles.detail.perm_timeout_members_desc",
				tier: "next",
			},
		],
	},
	{
		id: "text",
		titleKey: "group_settings.roles.detail.perm_group_text",
		items: [
			{
				id: "send_messages",
				titleKey: "group_settings.roles.detail.perm_send_messages",
				descriptionKey: "group_settings.roles.detail.perm_send_messages_desc",
				tier: "core",
			},
			{
				id: "send_messages_in_threads",
				titleKey: "group_settings.roles.detail.perm_send_messages_in_threads",
				descriptionKey: "group_settings.roles.detail.perm_send_messages_in_threads_desc",
				tier: "next",
			},
			{
				id: "create_public_threads",
				titleKey: "group_settings.roles.detail.perm_create_public_threads",
				descriptionKey: "group_settings.roles.detail.perm_create_public_threads_desc",
				tier: "next",
			},
			{
				id: "create_private_threads",
				titleKey: "group_settings.roles.detail.perm_create_private_threads",
				descriptionKey: "group_settings.roles.detail.perm_create_private_threads_desc",
				tier: "next",
			},
			{
				id: "embed_links",
				titleKey: "group_settings.roles.detail.perm_embed_links",
				descriptionKey: "group_settings.roles.detail.perm_embed_links_desc",
				tier: "next",
			},
			{
				id: "attach_files",
				titleKey: "group_settings.roles.detail.perm_attach_files",
				descriptionKey: "group_settings.roles.detail.perm_attach_files_desc",
				tier: "core",
			},
			{
				id: "add_reactions",
				titleKey: "group_settings.roles.detail.perm_add_reactions",
				descriptionKey: "group_settings.roles.detail.perm_add_reactions_desc",
				tier: "core",
			},
			{
				id: "mention_everyone",
				titleKey: "group_settings.roles.detail.perm_mention_everyone",
				descriptionKey: "group_settings.roles.detail.perm_mention_everyone_desc",
				tier: "core",
			},
			{
				id: "manage_messages",
				titleKey: "group_settings.roles.detail.perm_manage_messages",
				descriptionKey: "group_settings.roles.detail.perm_manage_messages_desc",
				tier: "core",
			},
			{
				id: "pin_messages",
				titleKey: "group_settings.roles.detail.perm_pin_messages",
				descriptionKey: "group_settings.roles.detail.perm_pin_messages_desc",
				tier: "core",
			},
			{
				id: "bypass_slowmode",
				titleKey: "group_settings.roles.detail.perm_bypass_slowmode",
				descriptionKey: "group_settings.roles.detail.perm_bypass_slowmode_desc",
				tier: "next",
			},
			{
				id: "manage_threads",
				titleKey: "group_settings.roles.detail.perm_manage_threads",
				descriptionKey: "group_settings.roles.detail.perm_manage_threads_desc",
				tier: "next",
			},
			{
				id: "read_message_history",
				titleKey: "group_settings.roles.detail.perm_read_message_history",
				descriptionKey: "group_settings.roles.detail.perm_read_message_history_desc",
				tier: "core",
			},
			{
				id: "send_voice_messages",
				titleKey: "group_settings.roles.detail.perm_send_voice_messages",
				descriptionKey: "group_settings.roles.detail.perm_send_voice_messages_desc",
				tier: "next",
			},
			{
				id: "create_polls",
				titleKey: "group_settings.roles.detail.perm_create_polls",
				descriptionKey: "group_settings.roles.detail.perm_create_polls_desc",
				tier: "next",
			},
		],
	},
	{
		id: "voice",
		titleKey: "group_settings.roles.detail.perm_group_voice",
		items: [
			{
				id: "connect",
				titleKey: "group_settings.roles.detail.perm_connect",
				descriptionKey: "group_settings.roles.detail.perm_connect_desc",
				tier: "core",
			},
			{
				id: "speak",
				titleKey: "group_settings.roles.detail.perm_speak",
				descriptionKey: "group_settings.roles.detail.perm_speak_desc",
				tier: "core",
			},
			{
				id: "video",
				titleKey: "group_settings.roles.detail.perm_video",
				descriptionKey: "group_settings.roles.detail.perm_video_desc",
				tier: "core",
			},
			{
				id: "use_voice_activity",
				titleKey: "group_settings.roles.detail.perm_use_voice_activity",
				descriptionKey: "group_settings.roles.detail.perm_use_voice_activity_desc",
				tier: "next",
			},
			{
				id: "use_soundboard",
				titleKey: "group_settings.roles.detail.perm_use_soundboard",
				descriptionKey: "group_settings.roles.detail.perm_use_soundboard_desc",
				tier: "next",
			},
			{
				id: "priority_speaker",
				titleKey: "group_settings.roles.detail.perm_priority_speaker",
				descriptionKey: "group_settings.roles.detail.perm_priority_speaker_desc",
				tier: "next",
			},
			{
				id: "mute_members",
				titleKey: "group_settings.roles.detail.perm_mute_members",
				descriptionKey: "group_settings.roles.detail.perm_mute_members_desc",
				tier: "core",
			},
			{
				id: "deafen_members",
				titleKey: "group_settings.roles.detail.perm_deafen_members",
				descriptionKey: "group_settings.roles.detail.perm_deafen_members_desc",
				tier: "core",
			},
			{
				id: "move_members",
				titleKey: "group_settings.roles.detail.perm_move_members",
				descriptionKey: "group_settings.roles.detail.perm_move_members_desc",
				tier: "next",
			},
		],
	},
	{
		id: "plugins",
		titleKey: "group_settings.roles.detail.perm_group_plugins",
		items: [
			{
				id: "manage_plugins",
				titleKey: "group_settings.roles.detail.perm_manage_plugins",
				descriptionKey: "group_settings.roles.detail.perm_manage_plugins_desc",
				tier: "core",
			},
			{
				id: "use_plugins",
				titleKey: "group_settings.roles.detail.perm_use_plugins",
				descriptionKey: "group_settings.roles.detail.perm_use_plugins_desc",
				tier: "core",
			},
			{
				id: "create_events",
				titleKey: "group_settings.roles.detail.perm_create_events",
				descriptionKey: "group_settings.roles.detail.perm_create_events_desc",
				tier: "next",
			},
			{
				id: "manage_events",
				titleKey: "group_settings.roles.detail.perm_manage_events",
				descriptionKey: "group_settings.roles.detail.perm_manage_events_desc",
				tier: "next",
			},
		],
	},
	{
		id: "advanced",
		titleKey: "group_settings.roles.detail.perm_group_advanced",
		items: [
			{
				id: "administrator",
				titleKey: "group_settings.roles.detail.perm_administrator",
				descriptionKey: "group_settings.roles.detail.perm_administrator_desc",
				tier: "core",
				dangerous: true,
			},
		],
	},
];

export function defaultPermissionsForRole(): Record<string, boolean> {
	return {
		view_channels: true,
		manage_channels: false,
		manage_roles: false,
		manage_group: false,
		create_expressions: false,
		manage_expressions: false,
		view_audit_log: false,
		view_group_insights: false,
		manage_webhooks: false,
		create_invite: true,
		kick_members: false,
		ban_members: false,
		timeout_members: false,
		send_messages: true,
		send_messages_in_threads: true,
		create_public_threads: true,
		create_private_threads: true,
		embed_links: true,
		attach_files: true,
		add_reactions: true,
		mention_everyone: false,
		manage_messages: false,
		pin_messages: true,
		bypass_slowmode: false,
		manage_threads: false,
		read_message_history: true,
		send_voice_messages: true,
		create_polls: true,
		connect: true,
		speak: true,
		video: true,
		use_voice_activity: true,
		use_soundboard: false,
		priority_speaker: false,
		mute_members: false,
		deafen_members: false,
		move_members: false,
		manage_plugins: false,
		use_plugins: true,
		create_events: false,
		manage_events: false,
		administrator: false,
	};
}

export function roleDisplayTitle(role: GroupSettingsRoleRow, t: (key: string) => string): string {
	if (role.id === "everyone" || role.name === "everyone") {
		return t("group_settings.roles.default_role_title");
	}
	if (role.isSystem) {
		return t(`invite.role_${role.name}`);
	}
	return role.name;
}

export function defaultEveryoneRoleRow(memberTotal: number): GroupSettingsRoleRow {
	return {
		id: "everyone",
		name: "everyone",
		memberCount: memberTotal,
		isSystem: true,
		color: DEFAULT_ROLE_COLOR,
	};
}

export function builtinRoleRows(memberTotal: number): GroupSettingsRoleRow[] {
	const members = Math.max(0, memberTotal - 1);
	return [
		{
			id: "member",
			name: "member",
			memberCount: members,
			isSystem: true,
			color: "#3b82f6",
		},
	];
}

export function matchesRoleSearch(
	row: GroupSettingsRoleRow,
	query: string,
	roleLabel: string,
): boolean {
	const q = query.trim().toLowerCase();
	if (!q) {
		return true;
	}
	return (
		row.name.toLowerCase().includes(q) ||
		row.id.toLowerCase().includes(q) ||
		roleLabel.toLowerCase().includes(q)
	);
}
