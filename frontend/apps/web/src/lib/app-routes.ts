/** Signed-in hub when the user has no group membership (create / join). */
export const APP_HOME_PATH = "/home";

export type QuickViewId = "mentions" | "discover" | "drafts" | "members";

export const QUICK_VIEW_IDS: readonly QuickViewId[] = [
	"mentions",
	"discover",
	"drafts",
	"members",
] as const;

export function isQuickViewId(value: string): value is QuickViewId {
	return (QUICK_VIEW_IDS as readonly string[]).includes(value);
}

/** groupId and channelId are 24-char public ids (see publicEntityIdSchema in @wichat/contracts). */
export function channelPath(groupId: string, channelId: string) {
	return `/channel/${encodeURIComponent(groupId)}/${encodeURIComponent(channelId)}`;
}

export function quickViewPath(groupId: string, view: QuickViewId) {
	return `/channel/${encodeURIComponent(groupId)}/view/${view}`;
}
