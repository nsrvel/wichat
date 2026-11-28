"use client";

import { type ReactNode } from "react";
import { MentionMemberChip } from "@/features/chat/components/mention-member-chip";
import { MENTION_USER_ID_SOURCE } from "@/features/chat/lib/mention-chip-styles";
import { extractMentionPublicUserIds } from "@/features/chat/lib/mention-markdown";

// Leading `[` must be escaped — a bare `[@` would start a regex character class (String.raw does not escape it).
const MENTION_LINK_BODY_SOURCE = `\\[@([^\\]]+)\\]\\(wichat-mention:(${MENTION_USER_ID_SOURCE})\\)`;

function mentionLinkBodyPattern(): RegExp {
	return new RegExp(MENTION_LINK_BODY_SOURCE, "gi");
}

export function segmentMessageMentionText(
	text: string,
	selfUserId: string | undefined,
	groupId?: string,
	onMentionMemberClick?: (userId: string) => void,
): ReactNode[] {
	const nodes: ReactNode[] = [];
	let lastIndex = 0;

	for (const match of text.matchAll(mentionLinkBodyPattern())) {
		const index = match.index ?? 0;
		if (index > lastIndex) {
			nodes.push(<span key={`t-${lastIndex}`}>{text.slice(lastIndex, index)}</span>);
		}
		const label = (match[1] ?? "").replace(/^@/, "");
		const userId = match[2] ?? "";
		nodes.push(
			<MentionMemberChip
				key={`m-${index}`}
				userId={userId}
				label={label}
				selfUserId={selfUserId}
				groupId={groupId}
				onOpenProfile={onMentionMemberClick}
			/>,
		);
		lastIndex = index + match[0].length;
	}

	if (lastIndex < text.length) {
		nodes.push(<span key={`t-${lastIndex}`}>{text.slice(lastIndex)}</span>);
	}

	return nodes;
}

export function hasRenderableMentionSegments(text: string): boolean {
	return new RegExp(MENTION_LINK_BODY_SOURCE, "i").test(text);
}

export function MentionText({
	text,
	selfUserId,
	groupId,
	onMentionMemberClick,
	className,
}: {
	text: string;
	selfUserId?: string;
	groupId?: string;
	onMentionMemberClick?: (userId: string) => void;
	className?: string;
}) {
	if (!hasRenderableMentionSegments(text)) {
		return <>{text}</>;
	}
	const nodes = segmentMessageMentionText(text, selfUserId, groupId, onMentionMemberClick);
	if (nodes.length === 0) {
		return <>{text}</>;
	}
	return <span className={className}>{nodes}</span>;
}

export function messageMentionsSelf(
	mentionedUserIds: string[] | undefined,
	body: string,
	selfUserId: string | undefined,
): boolean {
	const selfId = selfUserId?.trim();
	if (!selfId) {
		return false;
	}
	if (mentionedUserIds?.some((id) => id.trim() === selfId)) {
		return true;
	}
	return extractMentionPublicUserIds(body).includes(selfId);
}
