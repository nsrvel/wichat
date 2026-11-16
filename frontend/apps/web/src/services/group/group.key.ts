export const Q_MY_GROUPS = "groups.mine";

export const groupBanKeys = {
	all: (groupId: string) => ["groups", groupId, "bans"] as const,
	list: (groupId: string, search: string, cursor: string) =>
		["groups", groupId, "bans", search, cursor] as const,
};

export const groupAuditLogKeys = {
	all: (groupId: string) => ["groups", groupId, "audit-logs"] as const,
	list: (groupId: string, actorId: string, action: string) =>
		["groups", groupId, "audit-logs", actorId, action] as const,
};
