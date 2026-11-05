/** Static chat shell mock (fixtures + ChannelView). Not wired to real channels. */
export const UI_SAMPLE_CHAT_PATH = "/dev/chat-ui";

export function isUiSampleChatPath(pathname: string | null | undefined) {
	if (!pathname) return false;
	return pathname === UI_SAMPLE_CHAT_PATH || pathname.startsWith(`${UI_SAMPLE_CHAT_PATH}/`);
}
