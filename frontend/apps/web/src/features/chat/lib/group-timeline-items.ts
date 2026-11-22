import type { ChannelMessage } from "@/features/chat/lib/channel-message";

export type TimelineItem =
	| {
			kind: "date-divider";
			key: string;
			dateKey: string;
			isoDate: string;
	  }
	| {
			kind: "message";
			key: string;
			message: ChannelMessage;
			compact: boolean;
	  };

const GROUP_WINDOW_MS = 5 * 60 * 1000;

function toLocalDayKey(iso: string): string {
	const d = new Date(iso);
	if (Number.isNaN(d.getTime())) {
		return "unknown";
	}
	return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}

export function groupTimelineItems(messages: ChannelMessage[]): TimelineItem[] {
	const items: TimelineItem[] = [];
	let prevDayKey = "";
	let prevMsg: ChannelMessage | null = null;

	for (const msg of messages) {
		const dayKey = msg.dayKey ?? toLocalDayKey(msg.createdAt);
		if (dayKey !== prevDayKey) {
			items.push({
				kind: "date-divider",
				key: `day-${dayKey}`,
				dateKey: dayKey,
				isoDate: msg.createdAt,
			});
			prevDayKey = dayKey;
			prevMsg = null;
		}

		let compact = false;
		if (prevMsg && !msg.replyTo && prevMsg.authorId === msg.authorId) {
			const prevTime = prevMsg.createdAtMs ?? new Date(prevMsg.createdAt).getTime();
			const currTime = msg.createdAtMs ?? new Date(msg.createdAt).getTime();
			if (
				!Number.isNaN(prevTime) &&
				!Number.isNaN(currTime) &&
				currTime - prevTime <= GROUP_WINDOW_MS
			) {
				compact = true;
			}
		}

		const stableKey = msg.clientMsgId ? `cm:${msg.clientMsgId}` : msg.id;
		items.push({
			kind: "message",
			key: stableKey,
			message: msg,
			compact,
		});
		prevMsg = msg;
	}

	return items;
}
