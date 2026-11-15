/** Opt-in islands for WiChat context menus (Discord-style: browser menu off elsewhere). */
export const WICHAT_CONTEXT_MENU_ATTR = "data-wichat-context-menu";

/** Opt-in islands for selectable content/identifiers inside the select-none app shell. */
export const WICHAT_SELECTABLE_ATTR = "data-wichat-selectable";

export function isNativeContextMenuAllowed(target: EventTarget | null): boolean {
	if (!(target instanceof Element)) {
		return false;
	}
	if (target.closest(`[${WICHAT_CONTEXT_MENU_ATTR}]`)) {
		return true;
	}
	if (target.closest('input, textarea, [contenteditable="true"]')) {
		return true;
	}
	if (
		target.closest(`pre, code, [${WICHAT_SELECTABLE_ATTR}="true"]`) &&
		typeof window !== "undefined" &&
		Boolean(window.getSelection()?.toString().trim())
	) {
		return true;
	}
	return false;
}

export function suppressDefaultContextMenuUnlessAllowed(event: {
	target: EventTarget | null;
	preventDefault: () => void;
}) {
	if (!isNativeContextMenuAllowed(event.target)) {
		event.preventDefault();
	}
}
