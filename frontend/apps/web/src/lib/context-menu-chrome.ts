/** Opt-in islands for WiChat context menus (Discord-style: browser menu off elsewhere). */
export const WICHAT_CONTEXT_MENU_ATTR = "data-wichat-context-menu";

export function isNativeContextMenuAllowed(target: EventTarget | null): boolean {
	if (!(target instanceof Element)) {
		return false;
	}
	if (target.closest(`[${WICHAT_CONTEXT_MENU_ATTR}]`)) {
		return true;
	}
	return Boolean(target.closest('input, textarea, [contenteditable="true"]'));
}

export function suppressDefaultContextMenuUnlessAllowed(event: {
	target: EventTarget | null;
	preventDefault: () => void;
}) {
	if (!isNativeContextMenuAllowed(event.target)) {
		event.preventDefault();
	}
}
