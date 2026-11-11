"use client";

import { createContext, useContext, type ReactNode } from "react";
import { cn } from "@/lib/utils";

function escapeRegExp(value: string) {
	return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

export function buildMentionPattern(handles: readonly string[]) {
	const labels = [...handles].sort((a, b) => b.length - a.length).map(escapeRegExp);
	return new RegExp(`@(?:${labels.join("|")})`, "g");
}

type MentionTone = "incoming" | "outgoing";

export function MentionText({
	text,
	tone = "incoming",
	pattern,
}: {
	text: string;
	tone?: MentionTone;
	pattern: RegExp;
}) {
	const mentionClass =
		tone === "outgoing"
			? "font-semibold text-chat-mention-outgoing"
			: "font-semibold text-chat-mention";

	const nodes: ReactNode[] = [];
	let lastIndex = 0;
	for (const match of text.matchAll(pattern)) {
		const index = match.index ?? 0;
		if (index > lastIndex) {
			nodes.push(<span key={`t-${lastIndex}`}>{text.slice(lastIndex, index)}</span>);
		}
		const mention = match[0];
		nodes.push(
			<span key={`m-${index}`} className={cn(mentionClass, "cursor-pointer hover:underline")}>
				{mention}
			</span>,
		);
		lastIndex = index + mention.length;
	}
	if (lastIndex < text.length) {
		nodes.push(<span key={`t-${lastIndex}`}>{text.slice(lastIndex)}</span>);
	}
	if (nodes.length === 0) {
		return <>{text}</>;
	}
	return <>{nodes}</>;
}

const MentionToneContext = createContext<MentionTone>("incoming");

export function MentionToneProvider({
	tone,
	children,
}: {
	tone: MentionTone;
	children: ReactNode;
}) {
	return <MentionToneContext.Provider value={tone}>{children}</MentionToneContext.Provider>;
}

export function MessageMentionText({ text, pattern }: { text: string; pattern: RegExp }) {
	const tone = useContext(MentionToneContext);
	return <MentionText text={text} tone={tone} pattern={pattern} />;
}
