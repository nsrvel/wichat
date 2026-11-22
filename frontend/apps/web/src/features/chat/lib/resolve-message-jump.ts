import { findMessageSeqInCache, type ChannelMessagesInfiniteData } from "@/features/chat/lib/message-cache-window";
import { resolveMessageChannelSeq } from "@/services/message/message.service";

export type ResolveJumpSeqInput = {
	groupId: string;
	channelId: string;
	messageId: string;
	cache: ChannelMessagesInfiniteData | undefined;
	signal?: AbortSignal;
};

/** Prefer cache hit; otherwise resolve message id → channel_seq via API. */
export async function resolveJumpSeq(input: ResolveJumpSeqInput): Promise<number> {
	const cached = findMessageSeqInCache(input.cache, input.messageId);
	if (cached > 0) {
		return cached;
	}
	const resolved = await resolveMessageChannelSeq(
		input.groupId,
		input.channelId,
		input.messageId,
		input.signal,
	);
	return resolved.channel_seq;
}