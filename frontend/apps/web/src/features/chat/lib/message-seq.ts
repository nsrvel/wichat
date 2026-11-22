import type { Message } from "@wichat/contracts/chat/message";

export function parseChannelSeq(seq: string | undefined): number {
	if (!seq || !/^\d+$/.test(seq)) {
		return 0;
	}
	const n = Number(seq);
	return Number.isSafeInteger(n) ? n : 0;
}

export function compareChannelSeq(a: string | undefined, b: string | undefined): number {
	return parseChannelSeq(a) - parseChannelSeq(b);
}

export function highestChannelSeq(messages: Message[]): number {
	let max = 0;
	for (const msg of messages) {
		const seq = parseChannelSeq(msg.channel_seq);
		if (seq > max) {
			max = seq;
		}
	}
	return max;
}

export function isPendingMessageId(id: string): boolean {
	return id.startsWith("pending_");
}

export function newClientMessageId(): string {
	if (typeof crypto !== "undefined" && crypto.randomUUID) {
		return crypto.randomUUID();
	}
	return `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
}