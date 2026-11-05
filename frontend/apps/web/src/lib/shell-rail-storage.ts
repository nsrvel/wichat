const STORAGE_KEY = "wichat.shell.rail_open";

const DEFAULT_RAIL_OPEN = true;

export function readShellRailOpen(): boolean {
	if (typeof window === "undefined") {
		return DEFAULT_RAIL_OPEN;
	}
	try {
		const raw = window.localStorage.getItem(STORAGE_KEY);
		if (raw === "0" || raw === "false") {
			return false;
		}
		if (raw === "1" || raw === "true") {
			return true;
		}
	} catch {
		// private mode / blocked storage
	}
	return DEFAULT_RAIL_OPEN;
}

export function writeShellRailOpen(open: boolean) {
	if (typeof window === "undefined") {
		return;
	}
	try {
		window.localStorage.setItem(STORAGE_KEY, open ? "1" : "0");
	} catch {
		// ignore
	}
}
