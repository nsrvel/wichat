const EVENT_NAME = "wichat:jump-to-message";

export type JumpToMessageDetail = {
	messageId: string;
};

export function dispatchJumpToMessage(messageId: string): void {
	if (typeof window === "undefined") {
		return;
	}
	window.dispatchEvent(
		new CustomEvent<JumpToMessageDetail>(EVENT_NAME, {
			detail: { messageId },
		}),
	);
}

export function subscribeJumpToMessage(handler: (messageId: string) => void): () => void {
	if (typeof window === "undefined") {
		return () => {};
	}
	const listener = (event: Event) => {
		const detail = (event as CustomEvent<JumpToMessageDetail>).detail;
		if (detail?.messageId) {
			handler(detail.messageId);
		}
	};
	window.addEventListener(EVENT_NAME, listener);
	return () => window.removeEventListener(EVENT_NAME, listener);
}
