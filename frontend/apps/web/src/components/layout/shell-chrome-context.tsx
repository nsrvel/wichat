"use client";

import { createContext, useContext, useMemo, useState, type ReactNode } from "react";
import { CHANNEL_SIDEBAR_MIN_WIDTH } from "@/hooks/use-resizable-sidebar-width";

/** Icon rail column — matches `w-16` (64px). */
export const ICON_RAIL_WIDTH_PX = 64;

type ShellChromeContextValue = {
	railOpen: boolean;
	sidebarWidth: number;
	setSidebarWidth: (width: number) => void;
	voiceConnected: boolean;
	setVoiceConnected: (connected: boolean) => void;
	/** Sidebar filter chevron / extra row open — chat pinned bar uses zen column when true. */
	sidebarFilterExpanded: boolean;
	setSidebarFilterExpanded: (expanded: boolean) => void;
};

const ShellChromeContext = createContext<ShellChromeContextValue | null>(null);

type ShellChromeProviderProps = {
	railOpen: boolean;
	children: ReactNode;
};

export function ShellChromeProvider({ railOpen, children }: ShellChromeProviderProps) {
	const [sidebarWidth, setSidebarWidth] = useState(CHANNEL_SIDEBAR_MIN_WIDTH);
	const [voiceConnected, setVoiceConnected] = useState(true);
	const [sidebarFilterExpanded, setSidebarFilterExpanded] = useState(false);

	const value = useMemo(
		() => ({
			railOpen,
			sidebarWidth,
			setSidebarWidth,
			voiceConnected,
			setVoiceConnected,
			sidebarFilterExpanded,
			setSidebarFilterExpanded,
		}),
		[railOpen, sidebarWidth, voiceConnected, sidebarFilterExpanded],
	);

	return <ShellChromeContext.Provider value={value}>{children}</ShellChromeContext.Provider>;
}

export function useShellChrome() {
	const ctx = useContext(ShellChromeContext);
	if (!ctx) {
		throw new Error("useShellChrome must be used within ShellChromeProvider");
	}
	return ctx;
}

/** Dock card height + outer `p-2` on ShellUserDock. */
const DOCK_CHROME_PADDING = 16;

export function useShellChromeDockInset() {
	const { voiceConnected } = useShellChrome();
	return (voiceConnected ? 172 : 56) + DOCK_CHROME_PADDING;
}
