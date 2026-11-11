import type { InvitePreview } from "@wichat/contracts";
import { InvitePreviewSchema } from "@wichat/contracts";

const PREFIX = "wichat:invite-preview:";

function storageKey(token: string): string {
	return `${PREFIX}${token.trim()}`;
}

export function readInvitePreviewCache(token: string): InvitePreview | undefined {
	if (typeof sessionStorage === "undefined" || !token.trim()) {
		return undefined;
	}
	try {
		const raw = sessionStorage.getItem(storageKey(token));
		if (!raw) {
			return undefined;
		}
		const parsed = InvitePreviewSchema.safeParse(JSON.parse(raw));
		return parsed.success ? parsed.data : undefined;
	} catch {
		return undefined;
	}
}

export function writeInvitePreviewCache(token: string, preview: InvitePreview): void {
	if (typeof sessionStorage === "undefined" || !token.trim()) {
		return;
	}
	try {
		sessionStorage.setItem(storageKey(token), JSON.stringify(preview));
	} catch {
		// quota / private mode
	}
}
