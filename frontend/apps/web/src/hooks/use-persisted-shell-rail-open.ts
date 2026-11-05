"use client";

import { useCallback, useState } from "react";
import { readShellRailOpen, writeShellRailOpen } from "@/lib/shell-rail-storage";

/** Shell group rail expanded/collapsed — restored from `localStorage` after mount. */
export function usePersistedShellRailOpen() {
	const [railOpen, setRailOpenState] = useState(() => readShellRailOpen());

	const setRailOpen = useCallback((next: boolean | ((prev: boolean) => boolean)) => {
		setRailOpenState((prev) => {
			const value = typeof next === "function" ? next(prev) : next;
			writeShellRailOpen(value);
			return value;
		});
	}, []);

	return [railOpen, setRailOpen] as const;
}
