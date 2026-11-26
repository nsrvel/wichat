import {
	mentionChipClassName,
	MENTION_USER_ID_SOURCE,
} from "@/features/chat/lib/mention-chip-styles";

/** Stored user mention: `[@Label](wichat-mention:USER_ID)` */
export const MENTION_LINK_PREFIX = "wichat-mention:";

const MENTION_LINK_RE = new RegExp(
	`\\[@([^\\]]+)\\]\\(${MENTION_LINK_PREFIX}(${MENTION_USER_ID_SOURCE})\\)`,
	"gi",
);

export function extractMentionPublicUserIds(body: string): string[] {
	const seen = new Set<string>();
	const out: string[] = [];
	for (const match of body.matchAll(MENTION_LINK_RE)) {
		const id = match[2]?.trim();
		if (!id || seen.has(id)) {
			continue;
		}
		seen.add(id);
		out.push(id);
	}
	return out;
}

export function messageBodyMentionsSelf(body: string, selfUserId: string | undefined): boolean {
	if (!selfUserId?.trim()) {
		return false;
	}
	return extractMentionPublicUserIds(body).includes(selfUserId.trim());
}

export function isMentionLinkHref(href: string | undefined): boolean {
	return Boolean(href?.startsWith(MENTION_LINK_PREFIX));
}

export function mentionUserIdFromHref(href: string): string {
	return href.slice(MENTION_LINK_PREFIX.length).trim();
}

export function mentionLinkClassName(href: string | undefined, selfUserId?: string): string {
	if (!isMentionLinkHref(href)) {
		return "";
	}
	return mentionChipClassName(mentionUserIdFromHref(href ?? ""), selfUserId);
}