import { isPublicEntityId } from "@wichat/contracts";
import { isQuickViewId, type QuickViewId } from "@/lib/app-routes";

export type WorkspacePathTarget =
	| { channelId: string; quickView?: undefined }
	| { quickView: QuickViewId; channelId?: undefined }
	| { channelId?: undefined; quickView?: undefined };

/** Derive workspace shell target from a `/channel/{groupId}/…` pathname. */
export function parseWorkspacePath(pathname: string, groupId: string): WorkspacePathTarget {
	const prefix = `/channel/${groupId}/`;
	if (!pathname.startsWith(prefix)) {
		return {};
	}

	const rest = pathname.slice(prefix.length);
	if (rest.startsWith("view/")) {
		const segment = rest.slice("view/".length).split("/")[0] ?? "";
		if (isQuickViewId(segment)) {
			return { quickView: segment };
		}

		return {};
	}

	const segment = rest.split("/")[0] ?? "";
	if (isPublicEntityId(segment)) {
		return { channelId: segment };
	}

	return {};
}
