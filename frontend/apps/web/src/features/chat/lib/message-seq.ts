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

function randomUUIDv4(): string {
	const cryptoObj = typeof crypto !== "undefined" ? crypto : undefined;
	if (cryptoObj?.randomUUID) {
		return cryptoObj.randomUUID();
	}
	if (cryptoObj?.getRandomValues) {
		const bytes = new Uint8Array(16);
		cryptoObj.getRandomValues(bytes);
		bytes[6] = (bytes[6] & 0x0f) | 0x40;
		bytes[8] = (bytes[8] & 0x3f) | 0x80;
		const hex = [...bytes].map((b) => b.toString(16).padStart(2, "0")).join("");
		return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
	}
	return `${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;
}

export function newClientMessageId(): string {
	return randomUUIDv4();
}