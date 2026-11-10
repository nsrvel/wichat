"use client";

import { Link2, Lock } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { WiChatDialogFooter } from "@/components/ui/wichat-dialog";
import { shareDialogFooterButtonClass } from "@/features/invite/lib/share-dialog-chrome";
import {
	authSubmitClass,
	authSubmitGhostClass,
} from "@/features/auth/components/auth-shell";
import { cn } from "@/lib/utils";
import type { RegistrationMode } from "@wichat/contracts";

type ShareDialogFooterProps = {
	currentMode: RegistrationMode | undefined;
	onCopyLink: () => void;
	copyDisabled: boolean;
	linkCopied?: boolean;
	showCopy: boolean;
};

export function ShareDialogFooter({
	currentMode,
	onCopyLink,
	copyDisabled,
	linkCopied = false,
	showCopy,
}: ShareDialogFooterProps) {
	const { t } = useTranslation();
	const mode = currentMode ?? "invite_only";

	const labelFor = (m: RegistrationMode) => {
		switch (m) {
			case "open":
				return t("invite.access_open");
			case "approval_required":
				return t("invite.access_approval");
			default:
				return t("invite.access_invite_only");
		}
	};

	return (
		<WiChatDialogFooter>
			<div className="flex items-center justify-between gap-3">
				<div
					className={cn(
						authSubmitGhostClass,
						shareDialogFooterButtonClass,
						"inline-flex items-center gap-2",
					)}
					aria-label={t("invite.access_policy_label", { mode: labelFor(mode) })}
				>
					<Lock className="h-4 w-4 shrink-0" aria-hidden />
					<span className="truncate">{labelFor(mode)}</span>
				</div>

				{showCopy ? (
					<Button
						type="button"
						className={cn(authSubmitClass, shareDialogFooterButtonClass, "gap-2")}
						disabled={copyDisabled}
						onClick={onCopyLink}
					>
						<Link2 className="h-4 w-4 shrink-0" aria-hidden />
						<span className="truncate">
							{linkCopied ? t("invite.copy_link_done") : t("invite.copy_link")}
						</span>
					</Button>
				) : (
					<span className="min-w-[7.5rem] shrink-0" aria-hidden />
				)}
			</div>
		</WiChatDialogFooter>
	);
}
