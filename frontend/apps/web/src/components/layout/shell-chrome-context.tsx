"use client";

import { createContext, useContext, useMemo, useState, type ReactNode } from "react";
import { SHELL_RAIL_WIDTH_PX } from "@/components/layout/shell-rail/shell-rail-metrics";
import { CHANNEL_SIDEBAR_MIN_WIDTH } from "@/hooks/use-resizable-sidebar-width";

/** Icon rail column width in px — derived from tile + insets. */
export const ICON_RAIL_WIDTH_PX = SHELL_RAIL_WIDTH_PX;

type ShellChromeContextValue = {
	railOpen: boolean;
	sidebarWidth: number;
	setSidebarWidth: (width: number) => void;
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
	const [sidebarFilterExpanded, setSidebarFilterExpanded] = useState(false);

	const value = useMemo(
		() => ({
			railOpen,
			sidebarWidth,
			setSidebarWidth,
			sidebarFilterExpanded,
			setSidebarFilterExpanded,
		}),
		[railOpen, sidebarWidth, sidebarFilterExpanded],
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
