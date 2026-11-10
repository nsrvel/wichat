import { resolveGroupLogoUrl } from "@/features/shell/lib/group-logo-url";

export type ResolveChannelIconInput = {
	type: string;
	imageUrl?: string;
};

const LEGACY_KIND: Record<string, string> = {
	public: "text_and_voice",
	private: "text_and_voice",
};

const KIND_DEFAULT_PRESET: Record<string, string> = {
	broadcast: "bell",
	text_and_voice: "chat",
	text_only: "chat",
	forum: "chat",
	voice_only: "mic",
};

function normalizeChannelKind(kind: string): string {
	const key = kind.trim().toLowerCase();
	return LEGACY_KIND[key] ?? key;
}

function presetAssetPath(key: string): string {
	return `/icons/channel-${key.trim().toLowerCase()}.png`;
}

function resolveImageUrlValue(raw: string): string {
	const trimmed = raw.trim();
	if (/^https?:\/\//i.test(trimmed)) {
		return trimmed;
	}
	if (trimmed.startsWith("//")) {
		return `https:${trimmed}`;
	}
	if (trimmed.startsWith("/icons/")) {
		return trimmed;
	}
	if (trimmed.startsWith("/")) {
		const resolved = resolveGroupLogoUrl(trimmed);
		return resolved || trimmed;
	}
	return presetAssetPath(trimmed);
}

/** Bundled preset or uploaded URL; empty image uses defaults from channel kind (`type`). */
export function resolveChannelIconSrc({
	type,
	imageUrl = "",
}: ResolveChannelIconInput): string | undefined {
	const trimmed = imageUrl.trim();
	if (trimmed !== "") {
		return resolveImageUrlValue(trimmed);
	}

	const kind = normalizeChannelKind(type);
	const key = KIND_DEFAULT_PRESET[kind];
	if (!key) {
		return undefined;
	}
	return presetAssetPath(key);
}

export function channelTileToneForType(channelKind: string): "default" | "announcement" {
	return normalizeChannelKind(channelKind) === "broadcast" ? "announcement" : "default";
}
