"use client";

import type { GroupMember } from "@wichat/contracts/group/member";
import { Check, EyeOff, Info } from "lucide-react";
import { useId } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { UserAvatar } from "@/components/ui/user-avatar";
import {
	WiChatDialog,
	WiChatDialogBody,
	WiChatDialogFooter,
	WiChatDialogFooterActions,
} from "@/components/ui/wichat-dialog";
import { authSubmitClass, authSubmitGhostClass } from "@/lib/form-field-styles";
import { userInitials } from "@/lib/user-initials";
import { cn } from "@/lib/utils";
import { useSetUserRelationshipMutation } from "@/services/user/user.mutation";

const relationshipDialogButtonClass = "w-auto min-w-[7.5rem] max-w-[55%] shrink-0 px-8";

type MemberIgnoreDialogProps = {
	member: GroupMember | null;
	onClose: () => void;
	onSwitchToBlock: (member: GroupMember) => void;
};

/**
 * Discord-style confirmation modal for ignoring a user, with an inline
 * "Block instead -> View details" escalation card and standard WiChat dialog footer actions.
 */
export function MemberIgnoreDialog({ member, onClose, onSwitchToBlock }: MemberIgnoreDialogProps) {
	const { t } = useTranslation();
	const titleId = useId();
	const setRelationship = useSetUserRelationshipMutation();

	const displayName = member?.display_name || member?.username || "";

	const handleConfirmIgnore = () => {
		if (!member) {
			return;
		}
		setRelationship.mutate(
			{ targetUserId: member.user_id, kind: "ignore" },
			{
				onSuccess: () => {
					toast.success(t("group_settings.members.ignore_dialog.success", { name: displayName }));
					onClose();
				},
				onError: () => {
					toast.error(t("error.unknown"));
				},
			},
		);
	};

	return (
		<WiChatDialog
			open={member !== null}
			onOpenChange={(open) => {
				if (!open && !setRelationship.isPending) {
					onClose();
				}
			}}
			titleId={titleId}
			elevated
			className="max-w-[440px]"
		>
			{member ? (
				<>
					<WiChatDialogBody className="space-y-5 pt-1">
						<div className="flex flex-col items-center text-center">
							<div className="relative mb-3">
								<UserAvatar
									src={member.avatar_url}
									initials={userInitials(displayName)}
									size="xl"
								/>
								<span className="absolute -bottom-0.5 -right-0.5 flex size-7 items-center justify-center rounded-full border-2 border-card bg-muted text-muted-foreground shadow-sm">
									<EyeOff className="size-3.5" aria-hidden />
								</span>
							</div>
							<h2 id={titleId} className="text-xl font-bold tracking-tight text-foreground">
								{t("group_settings.members.ignore_dialog.title", { name: displayName })}
							</h2>
							<p className="mt-1 text-sm text-muted-foreground">
								{t("group_settings.members.ignore_dialog.subtitle")}
							</p>
						</div>

						<div className="divide-y divide-border/60">
							<div className="flex items-start gap-3.5 py-3">
								<span className="mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-full bg-muted-foreground/20 text-foreground">
									<Check className="size-3.5" strokeWidth={2.5} aria-hidden />
								</span>
								<div className="min-w-0">
									<p className="text-sm font-semibold text-foreground">
										{t("group_settings.members.ignore_dialog.bullet_hide_title")}
									</p>
									<p className="text-xs text-muted-foreground">
										{t("group_settings.members.ignore_dialog.bullet_hide_desc")}
									</p>
								</div>
							</div>
							<div className="flex items-start gap-3.5 py-3">
								<span className="mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-full bg-muted-foreground/20 text-foreground">
									<Check className="size-3.5" strokeWidth={2.5} aria-hidden />
								</span>
								<div className="min-w-0">
									<p className="text-sm font-semibold text-foreground">
										{t("group_settings.members.ignore_dialog.bullet_mute_title")}
									</p>
									<p className="text-xs text-muted-foreground">
										{t("group_settings.members.ignore_dialog.bullet_mute_desc")}
									</p>
								</div>
							</div>
							<div className="flex items-start gap-3.5 py-3">
								<span className="mt-0.5 flex size-6 shrink-0 items-center justify-center rounded-full bg-muted-foreground/20 text-foreground">
									<Info className="size-3.5" strokeWidth={2.25} aria-hidden />
								</span>
								<div className="min-w-0">
									<p className="text-sm font-semibold text-foreground">
										{t("group_settings.members.ignore_dialog.bullet_info_title")}
									</p>
									<p className="text-xs text-muted-foreground">
										{t("group_settings.members.ignore_dialog.bullet_info_desc")}
									</p>
								</div>
							</div>
						</div>

						<div className="space-y-2">
							<p className="text-xs font-semibold text-foreground">
								{t("group_settings.members.ignore_dialog.not_enough")}
							</p>
							<div className="flex items-center justify-between gap-3 rounded-xl border border-border/60 bg-muted/50 px-3.5 py-3">
								<div className="min-w-0">
									<p className="text-sm font-semibold text-foreground">
										{t("group_settings.members.ignore_dialog.block_instead")}
									</p>
									<p className="text-xs text-muted-foreground">
										{t("group_settings.members.ignore_dialog.block_instead_desc")}
									</p>
								</div>
								<Button
									type="button"
									variant="secondary"
									size="sm"
									className="shrink-0"
									onClick={() => onSwitchToBlock(member)}
								>
									{t("group_settings.members.ignore_dialog.view_details")}
								</Button>
							</div>
						</div>
					</WiChatDialogBody>
					<WiChatDialogFooter>
						<WiChatDialogFooterActions
							backLabel={t("group_settings.members.ignore_dialog.cancel")}
							onBack={onClose}
							primaryLabel={t("group_settings.members.ignore_dialog.confirm")}
							onPrimary={handleConfirmIgnore}
							primaryDisabled={setRelationship.isPending}
							backClassName={cn(authSubmitGhostClass, relationshipDialogButtonClass)}
							primaryClassName={cn(authSubmitClass, relationshipDialogButtonClass)}
						/>
					</WiChatDialogFooter>
				</>
			) : null}
		</WiChatDialog>
	);
}
