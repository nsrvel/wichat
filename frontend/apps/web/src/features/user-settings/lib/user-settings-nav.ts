import {
	Accessibility,
	Bell,
	Code2,
	Globe,
	IdCard,
	Mic,
	MonitorSmartphone,
	Palette,
	Shield,
	UserRoundPen,
	type LucideIcon,
} from "lucide-react";

export type UserSettingsSectionId =
	| "account"
	| "profiles"
	| "privacy"
	| "notifications"
	| "appearance"
	| "voice"
	| "accessibility"
	| "language"
	| "devices"
	| "developer";

export type UserSettingsSubAnchor = {
	id: string;
	labelKey: string;
};

export type UserSettingsSearchField = {
	id: string;
	elementId: string;
	labelKey: string;
	keywords?: string[];
};

export type UserSettingsNavItem = {
	id: UserSettingsSectionId;
	labelKey: string;
	subtitleKey?: string;
	icon: LucideIcon;
	keywords: string[];
	subAnchors?: UserSettingsSubAnchor[];
	fields?: UserSettingsSearchField[];
};

export type UserSettingsNavGroup = {
	id: "user" | "app" | "advanced";
	labelKey: string;
	items: UserSettingsNavItem[];
};

export const USER_SETTINGS_NAV_GROUPS: UserSettingsNavGroup[] = [
	{
		id: "user",
		labelKey: "user_settings.groups.user",
		items: [
			{
				id: "profiles",
				labelKey: "user_settings.nav.profiles",
				subtitleKey: "user_settings.profiles.subtitle",
				icon: UserRoundPen,
				keywords: ["profile", "avatar", "banner", "bio", "name", "pronouns", "profil"],
				fields: [
					{
						id: "avatar",
						elementId: "user-settings-avatar",
						labelKey: "user_settings.profiles.avatar",
						keywords: ["photo", "foto", "picture"],
					},
					{
						id: "display-name",
						elementId: "user-settings-display-name",
						labelKey: "user_settings.profiles.display_name",
						keywords: ["name", "nama"],
					},
					{
						id: "pronouns",
						elementId: "user-settings-pronouns",
						labelKey: "user_settings.profiles.pronouns",
						keywords: ["pronoun", "he", "she", "they"],
					},
					{
						id: "bio",
						elementId: "user-settings-bio",
						labelKey: "user_settings.profiles.bio",
						keywords: ["bio", "about", "tentang"],
					},
					{
						id: "banner-color",
						elementId: "user-settings-banner-color",
						labelKey: "user_settings.profiles.banner_color",
						keywords: ["banner", "color", "warna", "hex"],
					},
				],
			},
			{
				id: "account",
				labelKey: "user_settings.nav.account",
				icon: IdCard,
				keywords: ["account", "username", "email", "password", "security", "mfa", "status", "akun"],
				subAnchors: [
					{ id: "account-info", labelKey: "user_settings.nav.account_info" },
					{ id: "password-security", labelKey: "user_settings.nav.password_security" },
					{ id: "account-status", labelKey: "user_settings.nav.account_status" },
				],
				fields: [
					{
						id: "username",
						elementId: "user-settings-account-info",
						labelKey: "user_settings.account.username",
						keywords: ["username", "handle"],
					},
					{
						id: "email",
						elementId: "user-settings-account-info",
						labelKey: "user_settings.account.email",
						keywords: ["email", "mail"],
					},
					{
						id: "password",
						elementId: "user-settings-password-security",
						labelKey: "user_settings.account.password",
						keywords: ["password", "kata sandi"],
					},
					{
						id: "mfa",
						elementId: "user-settings-password-security",
						labelKey: "user_settings.account.mfa",
						keywords: ["mfa", "2fa", "two-factor"],
					},
					{
						id: "devices",
						elementId: "user-settings-password-security",
						labelKey: "user_settings.account.devices",
						keywords: ["devices", "sessions", "perangkat"],
					},
					{
						id: "standing",
						elementId: "user-settings-account-status",
						labelKey: "user_settings.account.standing_title",
						keywords: ["standing", "status"],
					},
				],
			},
			{
				id: "privacy",
				labelKey: "user_settings.nav.privacy",
				icon: Shield,
				keywords: ["privacy", "data", "safety", "privasi"],
			},
			{
				id: "notifications",
				labelKey: "user_settings.nav.notifications",
				icon: Bell,
				keywords: ["notifications", "sound", "alerts", "notifikasi"],
			},
		],
	},
	{
		id: "app",
		labelKey: "user_settings.groups.app",
		items: [
			{
				id: "appearance",
				labelKey: "user_settings.nav.appearance",
				icon: Palette,
				keywords: ["appearance", "theme", "dark", "tampilan"],
			},
			{
				id: "voice",
				labelKey: "user_settings.nav.voice",
				icon: Mic,
				keywords: ["voice", "video", "mic", "audio", "suara"],
			},
			{
				id: "accessibility",
				labelKey: "user_settings.nav.accessibility",
				icon: Accessibility,
				keywords: ["accessibility", "motion", "font", "aksesibilitas"],
			},
			{
				id: "language",
				labelKey: "user_settings.nav.language",
				subtitleKey: "user_settings.language.subtitle",
				icon: Globe,
				keywords: ["language", "locale", "time", "english", "indonesia", "bahasa"],
				fields: [
					{
						id: "display-language",
						elementId: "user-settings-language-select",
						labelKey: "user_settings.language.select_label",
						keywords: ["english", "indonesia", "locale", "bahasa"],
					},
				],
			},
		],
	},
	{
		id: "advanced",
		labelKey: "user_settings.groups.advanced",
		items: [
			{
				id: "devices",
				labelKey: "user_settings.nav.devices",
				icon: MonitorSmartphone,
				keywords: ["devices", "sessions", "perangkat"],
			},
			{
				id: "developer",
				labelKey: "user_settings.nav.developer",
				icon: Code2,
				keywords: ["developer", "debug", "pengembang"],
			},
		],
	},
];

export const USER_SETTINGS_ITEM_BY_ID: Record<UserSettingsSectionId, UserSettingsNavItem> =
	Object.fromEntries(
		USER_SETTINGS_NAV_GROUPS.flatMap((group) => group.items.map((item) => [item.id, item])),
	) as Record<UserSettingsSectionId, UserSettingsNavItem>;

export const USER_SETTINGS_PLACEHOLDER_ICONS: Partial<Record<UserSettingsSectionId, LucideIcon>> = {
	privacy: Shield,
	notifications: Bell,
	appearance: Palette,
	voice: Mic,
	accessibility: Accessibility,
	devices: MonitorSmartphone,
	developer: Code2,
};
