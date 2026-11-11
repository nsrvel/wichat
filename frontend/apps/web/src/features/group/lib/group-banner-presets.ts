/** No solid banner — channel default chrome only. */
export const GROUP_BANNER_COLOR_DEFAULT = "default";

/** @deprecated Use GROUP_BANNER_COLOR_DEFAULT */
export const DEFAULT_GROUP_BANNER_COLOR = GROUP_BANNER_COLOR_DEFAULT;

/**
 * Discord-style banner palette (default first, then 5×2 grid).
 * IDs are stored in API/state; labels live in i18n (`banner_colors.*`).
 */
export const GROUP_BANNER_IMAGE_ID = "image";

/** Default + 8 colors + custom image (5×2 grid). */
export const GROUP_BANNER_COLOR_PRESETS = [
	"default",
	"blossom",
	"volcano",
	"sunset",
	"gold",
	"grape",
	"sky",
	"forest",
	"onyx",
	GROUP_BANNER_IMAGE_ID,
] as const;

export type GroupBannerColorId = (typeof GROUP_BANNER_COLOR_PRESETS)[number];

/** Swatch / preview surfaces — arbitrary values kept here for Tailwind scanning. */
export const GROUP_BANNER_SWATCH_SURFACE: Record<GroupBannerColorId, string> = {
	default: "bg-gradient-to-br from-[#5c6370] via-[#4f545c] to-[#2b2d31]",
	blossom: "bg-[#ff4cd2]",
	volcano: "bg-[#f23f43]",
	sunset: "bg-[#ff7518]",
	gold: "bg-[#fee75c]",
	grape: "bg-[#9c84ef]",
	sky: "bg-[#00a8fc]",
	forest: "bg-[#3ba55c]",
	onyx: "bg-[#23272a]",
	image: "bg-muted/50",
};

const PRESET_SET = new Set<string>(GROUP_BANNER_COLOR_PRESETS);

const LEGACY_ID_MAP: Record<string, GroupBannerColorId> = {
	default: "default",
	none: "default",
	"pink-500": "blossom",
	"red-500": "volcano",
	"orange-500": "sunset",
	"yellow-500": "gold",
	"violet-500": "grape",
	"purple-500": "grape",
	"blue-500": "sky",
	"teal-500": "forest",
	aqua: "forest",
	"green-500": "forest",
	image: GROUP_BANNER_IMAGE_ID,
	"neutral-900": "onyx",
	white: "default",
};

const HEX_TO_ID: Record<string, GroupBannerColorId> = {
	"#FF4CD2": "blossom",
	"#EC4899": "blossom",
	"#F472B6": "blossom",
	"#F23F43": "volcano",
	"#EF4444": "volcano",
	"#FF7518": "sunset",
	"#F97316": "sunset",
	"#FEE75C": "gold",
	"#EAB308": "gold",
	"#9C84EF": "grape",
	"#8B5CF6": "grape",
	"#A855F7": "grape",
	"#00A8FC": "sky",
	"#3B82F6": "sky",
	"#1CDC9C": "forest",
	"#14B8A6": "forest",
	"#3BA55C": "forest",
	"#22C55E": "forest",
	"#23272A": "onyx",
	"#171717": "onyx",
	"#141416": "onyx",
};

export function normalizeGroupBannerColor(raw: string | undefined): GroupBannerColorId {
	const trimmed = raw?.trim() ?? "";
	if (!trimmed) {
		return GROUP_BANNER_COLOR_DEFAULT;
	}
	const lower = trimmed.toLowerCase();
	if (PRESET_SET.has(lower)) {
		return lower as GroupBannerColorId;
	}
	if (LEGACY_ID_MAP[lower]) {
		return LEGACY_ID_MAP[lower];
	}
	if (trimmed.startsWith("#")) {
		const mapped = HEX_TO_ID[trimmed.toUpperCase()];
		if (mapped) {
			return mapped;
		}
	}
	return GROUP_BANNER_COLOR_DEFAULT;
}

export function isGroupBannerColorDefault(color: string | undefined): boolean {
	return normalizeGroupBannerColor(color) === GROUP_BANNER_COLOR_DEFAULT;
}

export function isGroupBannerImageId(id: string): id is typeof GROUP_BANNER_IMAGE_ID {
	return id === GROUP_BANNER_IMAGE_ID;
}

export function groupBannerSwatchSurfaceClass(color: string | undefined): string {
	const id = normalizeGroupBannerColor(color);
	if (isGroupBannerImageId(id)) {
		return GROUP_BANNER_SWATCH_SURFACE.image;
	}
	return GROUP_BANNER_SWATCH_SURFACE[id];
}

/** @deprecated Prefer groupBannerSwatchSurfaceClass */
export function groupBannerColorBgClass(color: string | undefined): string {
	return groupBannerSwatchSurfaceClass(color);
}

/** Selected swatch outline — neutral, matches form focus. */
export const groupBannerSwatchSelectedRingClass =
	"ring-2 ring-foreground ring-offset-2 ring-offset-background";

export const groupBannerColorSwatchClass =
	"relative aspect-[5/3] w-full overflow-hidden rounded-lg border border-border transition-transform hover:scale-[1.03]";

export type GroupBannerMode = "color" | "image";

// ponytail: 8 MiB cap for wide banners until ms-user documents a limit
export const MAX_GROUP_BANNER_BYTES = 8 * 1024 * 1024;
export const GROUP_BANNER_ACCEPT = "image/png,image/jpeg,image/webp";

/** Centered — matches sidebar card width (~284px). */
export const groupBannerPreviewWrapClass = "flex justify-center px-3 sm:px-5";

export const groupBannerPreviewStripClass =
	"relative h-[6.25rem] w-full max-w-[17.75rem] shrink-0 overflow-hidden rounded-lg border border-border";

export const groupProfileCardShellClass =
	"w-full max-w-[17.75rem] rounded-xl border border-border bg-card p-3 pb-4 shadow-sm";

export const groupProfileCardBodyClass = "flex flex-col items-center px-1 pt-2 text-center";

export const groupProfileCardIconClass =
	"-mt-7 size-[3.25rem] overflow-hidden rounded-full bg-field ring-[3px] ring-card shadow-sm";
