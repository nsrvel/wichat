export type GroupSettingsSection =
	"profile" | "onboarding" | "members" | "roles" | "audit_logs" | "bans";

export type GroupSettingsTarget = {
	groupId: string;
	groupName: string;
	logoUrl: string;
	description?: string;
	/** Server banner image URL when profile branding ships on the API. */
	bannerUrl?: string;
	/** Solid banner palette id (e.g. `volcano`) or omitted/`default` for no banner. */
	bannerColor?: string;
	/** When true, pre-join surfaces hide group profile details (invite / join). */
	isPrivate?: boolean;
	role?: string;
};

export type OpenGroupSettingsInput = GroupSettingsTarget & {
	section?: GroupSettingsSection;
};
