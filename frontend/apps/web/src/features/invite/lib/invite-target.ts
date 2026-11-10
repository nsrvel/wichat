export type InviteGroupTarget = {
	scope: "group";
	groupId: string;
	groupName: string;
};

export type InviteChannelTarget = {
	scope: "channel";
	groupId: string;
	groupName: string;
	channelId: string;
	channelName: string;
};

export type InviteTarget = InviteGroupTarget | InviteChannelTarget;
