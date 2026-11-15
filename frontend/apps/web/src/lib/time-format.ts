"use client";

import { useSyncExternalStore } from "react";

export type TimeFormatPreference = "auto" | "12h" | "24h";

const TIME_FORMAT_STORAGE_KEY = "wichat-time-format";
const TIME_FORMAT_EVENT = "wichat-time-format";

function isTimeFormatPreference(value: string | null): value is TimeFormatPreference {
	return value === "auto" || value === "12h" || value === "24h";
}

export function subscribeTimeFormat(onStoreChange: () => void): () => void {
	if (typeof window === "undefined") {
		return () => {};
	}
	const onStorage = (event: StorageEvent) => {
		if (event.key === null || event.key === TIME_FORMAT_STORAGE_KEY) {
			onStoreChange();
		}
	};
	window.addEventListener(TIME_FORMAT_EVENT, onStoreChange);
	window.addEventListener("storage", onStorage);
	return () => {
		window.removeEventListener(TIME_FORMAT_EVENT, onStoreChange);
		window.removeEventListener("storage", onStorage);
	};
}

export function readStoredTimeFormat(): TimeFormatPreference {
	try {
		const raw = window.localStorage.getItem(TIME_FORMAT_STORAGE_KEY);
		return isTimeFormatPreference(raw) ? raw : "auto";
	} catch {
		return "auto";
	}
}

export function writeStoredTimeFormat(value: TimeFormatPreference): void {
	try {
		window.localStorage.setItem(TIME_FORMAT_STORAGE_KEY, value);
		if (typeof window !== "undefined") {
			window.dispatchEvent(new Event(TIME_FORMAT_EVENT));
		}
	} catch {
		// Ignore storage failures in restricted browsing contexts.
	}
}

export function useTimeFormatPreference(): TimeFormatPreference {
	return useSyncExternalStore(
		subscribeTimeFormat,
		readStoredTimeFormat,
		() => "auto" as const,
	);
}

export function resolveHour12(preference: TimeFormatPreference): boolean | undefined {
	if (preference === "12h") return true;
	if (preference === "24h") return false;
	return undefined;
}
