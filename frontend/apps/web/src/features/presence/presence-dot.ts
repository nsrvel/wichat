import type { GroupMember } from "@wichat/contracts/group/group-member";
import type { MyPresenceResponse, PresenceStatus } from "@wichat/contracts/user/presence";
import type { UserPresenceStatus } from "@/lib/user-status";

export type AvatarPresenceDot = UserPresenceStatus | "offline";

/** Avatar dot: server effective only (WS-backed `Q_MY_PRESENCE`). */
export function presenceDotFrom(data: MyPresenceResponse | undefined): AvatarPresenceDot | undefined {
	if (!data) {
		return undefined;
	}
	return presenceDotFromFields(data.is_online, data.effective_status);
}

/** Member list / chat avatars: require `is_online` + `presence_status` from API + WS patches. */
export function memberPresenceDot(
	member: Pick<GroupMember, "is_online" | "presence_status"> | undefined,
): AvatarPresenceDot {
	if (!member) {
		return "offline";
	}
	return presenceDotFromFields(member.is_online ?? false, member.presence_status ?? "offline");
}

export function presenceDotFromFields(
	isOnline: boolean,
	effective: PresenceStatus | undefined,
): AvatarPresenceDot {
	if (!isOnline) {
		return "offline";
	}
	if (effective === "online" || effective === "idle" || effective === "dnd") {
		return effective;
	}
	return "offline";
}