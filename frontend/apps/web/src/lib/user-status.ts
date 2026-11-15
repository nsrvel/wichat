import type { User } from "@wichat/contracts";

export type UserPresenceStatus = "online" | "idle" | "dnd" | "invisible";

export type StatusClearDuration = "today" | "4h" | "1h" | "30m" | "never";

export type StatusNote = {
	text: string;
	emoji: string;
	expiresAt: number | null;
};

export const EMPTY_STATUS_NOTE: StatusNote = {
	text: "",
	emoji: "",
	expiresAt: null,
};

export type UserStatusMeta = {
	id: UserPresenceStatus;
	labelKey:
		| "shell.user_dock.status_online"
		| "shell.user_dock.status_idle"
		| "shell.user_dock.status_dnd"
		| "shell.user_dock.status_invisible";
	descriptionKey?:
		| "shell.user_dock.status_dnd_desc"
		| "shell.user_dock.status_invisible_desc";
	dotClass: string;
};

export const USER_STATUS_OPTIONS: readonly UserStatusMeta[] = [
	{
		id: "online",
		labelKey: "shell.user_dock.status_online",
		dotClass: "bg-emerald-500",
	},
	{
		id: "idle",
		labelKey: "shell.user_dock.status_idle",
		dotClass: "bg-amber-400",
	},
	{
		id: "dnd",
		labelKey: "shell.user_dock.status_dnd",
		descriptionKey: "shell.user_dock.status_dnd_desc",
		dotClass: "bg-rose-500",
	},
	{
		id: "invisible",
		labelKey: "shell.user_dock.status_invisible",
		descriptionKey: "shell.user_dock.status_invisible_desc",
		dotClass: "bg-muted-foreground/70",
	},
] as const;

export const USER_STATUS_BY_ID: Record<UserPresenceStatus, UserStatusMeta> = {
	online: USER_STATUS_OPTIONS[0],
	idle: USER_STATUS_OPTIONS[1],
	dnd: USER_STATUS_OPTIONS[2],
	invisible: USER_STATUS_OPTIONS[3],
};

const STATUS_STORAGE_KEY = "wichat-user-status";
const PRESENCE_EVENT = "wichat-user-presence";

function notifyPresenceChange(): void {
	if (typeof window !== "undefined") {
		window.dispatchEvent(new Event(PRESENCE_EVENT));
	}
}

export function subscribeUserPresence(onStoreChange: () => void): () => void {
	if (typeof window === "undefined") {
		return () => {};
	}
	const onStorage = (event: StorageEvent) => {
		if (event.key === null || event.key === STATUS_STORAGE_KEY) {
			onStoreChange();
		}
	};
	window.addEventListener(PRESENCE_EVENT, onStoreChange);
	window.addEventListener("storage", onStorage);
	return () => {
		window.removeEventListener(PRESENCE_EVENT, onStoreChange);
		window.removeEventListener("storage", onStorage);
	};
}

function isUserPresenceStatus(value: string | null): value is UserPresenceStatus {
	return value === "online" || value === "idle" || value === "dnd" || value === "invisible";
}

export function readStoredUserStatus(): UserPresenceStatus {
	try {
		const raw = window.localStorage.getItem(STATUS_STORAGE_KEY);
		return isUserPresenceStatus(raw) ? raw : "online";
	} catch {
		return "online";
	}
}

export function writeStoredUserStatus(status: UserPresenceStatus): void {
	try {
		window.localStorage.setItem(STATUS_STORAGE_KEY, status);
		notifyPresenceChange();
	} catch {
		// Ignore storage failures in restricted browsing contexts.
	}
}

export function statusNoteFromUser(
	user?: Pick<
		User,
		"status_note_text" | "status_note_emoji" | "status_note_expires_at"
	> | null,
): StatusNote {
	if (!user) return EMPTY_STATUS_NOTE;
	const text = user.status_note_text?.trim().slice(0, 80) ?? "";
	const emoji = user.status_note_emoji?.trim().slice(0, 16) ?? "";
	if (!text && !emoji) return EMPTY_STATUS_NOTE;

	let expiresAt: number | null = null;
	if (user.status_note_expires_at) {
		const parsed = Date.parse(user.status_note_expires_at);
		if (!Number.isNaN(parsed)) {
			if (Date.now() >= parsed) {
				return EMPTY_STATUS_NOTE;
			}
			expiresAt = parsed;
		}
	}
	return { text, emoji, expiresAt };
}

export function computeStatusNoteExpiry(duration: StatusClearDuration): number | null {
	const now = new Date();
	switch (duration) {
		case "30m":
			return now.getTime() + 30 * 60 * 1000;
		case "1h":
			return now.getTime() + 60 * 60 * 1000;
		case "4h":
			return now.getTime() + 4 * 60 * 60 * 1000;
		case "today": {
			const end = new Date(now);
			end.setHours(23, 59, 59, 999);
			return end.getTime();
		}
		case "never":
			return null;
	}
}

export function formatStatusNoteLabel(note: StatusNote): string {
	if (note.emoji && note.text) return `${note.emoji} ${note.text}`;
	return note.emoji || note.text;
}
