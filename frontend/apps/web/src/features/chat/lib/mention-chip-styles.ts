import { cn } from "@/lib/utils";

/** Display mention user id in stored markdown (24-hex entity id or UUID). */
export const MENTION_USER_ID_SOURCE = String.raw`(?:[0-9a-f]{24}|[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12})`;

export function mentionChipClassName(mentionUserId: string, selfUserId?: string): string {
	const selfId = selfUserId?.trim() ?? "";
	const id = mentionUserId.trim();
	const isSelf = Boolean(selfId) && id === selfId;
	return cn(
		"mention-chip inline rounded-md px-1 font-medium no-underline",
		isSelf
			? "bg-chat-mention-self-bg text-chat-mention-self"
			: "bg-[color-mix(in_oklab,var(--chat-mention-other)_15%,transparent)] text-chat-mention-other",
	);
}
