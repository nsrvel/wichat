export type MentionsFilterKind = "all" | "channels" | "chats";

export type MentionsFilterId = "unread" | MentionsFilterKind;

export function mentionsFilterToKind(
	filter: MentionsFilterId,
): MentionsFilterKind | "unread" | undefined {
	if (filter === "unread") {
		return "unread";
	}
	if (filter === "channels" || filter === "chats") {
		return filter;
	}
	return undefined;
}
