export const SHELL_SEARCH_FILTERS = ["messages", "files", "channels", "images"] as const;

export type ShellSearchFilterId = (typeof SHELL_SEARCH_FILTERS)[number];

export const DEFAULT_SHELL_SEARCH_FILTER: ShellSearchFilterId = "messages";

export function isShellSearchFilterEnabled(id: ShellSearchFilterId): boolean {
	return id === "messages" || id === "channels" || id === "files" || id === "images";
}
