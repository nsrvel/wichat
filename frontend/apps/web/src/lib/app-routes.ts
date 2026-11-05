/** Signed-in hub when the user has no group membership (create / join). */
export const APP_HOME_PATH = "/home";

/** groupId and channelId are 24-char public ids (see publicEntityIdSchema in @wichat/contracts). */
export function channelPath(groupId: string, channelId: string) {
	return `/channel/${encodeURIComponent(groupId)}/${encodeURIComponent(channelId)}`;
}
