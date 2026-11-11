import type { GroupsList } from "@wichat/contracts";

/** Fixture workspace for `/dev/chat-ui` only — not used on production routes. */
export const mockMyGroupsList: GroupsList = {
	groups: [
		{
			id: "f00100000000000000000001",
			name: "UI sample workspace",
			logo_url: "",
			description: "",
			is_private: false,
			banner_color: "",
			banner_url: "",
			default_channel_id: "f00100000000000000000002",
			role: "owner",
		},
	],
};

export const mockJoinableGroupsList: GroupsList = { groups: [] };
