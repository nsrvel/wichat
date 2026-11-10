"use client";

import { useTranslation } from "react-i18next";
import { WiChatDialogHeader } from "@/components/ui/wichat-dialog";
import type { InviteTarget } from "@/features/invite/lib/invite-target";

type ShareDialogHeaderProps = {
	target: InviteTarget;
	titleId: string;
	/** When false, subtitle is `invite.owner_only_share` (group invite policy — wire later). */
	canInviteAndCopy?: boolean;
};

export function ShareDialogHeader({
	target,
	titleId,
	canInviteAndCopy = true,
}: ShareDialogHeaderProps) {
	const { t } = useTranslation();

	const title =
		target.scope === "channel"
			? t("invite.header_title_channel", { channel: target.channelName })
			: t("invite.header_title_group");

	const shareLead =
		target.scope === "channel"
			? t("invite.dialog_lead_channel", {
					group: target.groupName,
					channel: target.channelName,
				})
			: t("invite.dialog_lead_group", { name: target.groupName });

	const subtitle = canInviteAndCopy ? shareLead : t("invite.owner_only_share");

	return <WiChatDialogHeader titleId={titleId} title={title} subtitle={subtitle} />;
}
