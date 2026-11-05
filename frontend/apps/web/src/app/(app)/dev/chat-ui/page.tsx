import type { Metadata } from "next";
import { DevChatUiShell } from "@/features/dev/chat-ui/dev-chat-ui-shell";
import { ChannelView } from "@/features/chat/components/channel-view";

export const metadata: Metadata = {
	title: "Chat UI sample — WiChat",
	description: "Static shell + mock workspace for visual reference while building production chat.",
};

/** Fixture-driven chat chrome; see fixtures/demo-group.ts and shell-navigation mocks. */
export default function DevChatUiSamplePage() {
	return (
		<DevChatUiShell>
			<ChannelView />
		</DevChatUiShell>
	);
}
