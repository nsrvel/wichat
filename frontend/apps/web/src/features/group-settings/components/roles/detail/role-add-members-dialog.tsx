"use client";

import type { GroupMember } from "@wichat/contracts/group/member";
import { Check, Search } from "lucide-react";
import { useId, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Input } from "@/components/ui/input";
import { OwnerCrownIcon } from "@/components/ui/owner-crown-icon";
import { UserAvatar } from "@/components/ui/user-avatar";
import {
	WiChatDialog,
	WiChatDialogBody,
	WiChatDialogFooter,
	WiChatDialogFooterActions,
	WiChatDialogHeader,
} from "@/components/ui/wichat-dialog";
import { groupSettingsToolbarSearchInputClass } from "@/features/group-settings/lib/group-settings-chrome";
import { ApiError } from "@wichat/api-client/api-error";
import {
	authFieldLabelClass,
	authFormFooterButtonClass,
	authInputClass,
	authSubmitClass,
	authSubmitGhostClass,
} from "@/lib/form-field-styles";
import { resolveMemberNameColor } from "@/lib/member-role-display";
import { userInitials } from "@/lib/user-initials";
import { cn } from "@/lib/utils";
import { useAddGroupRoleMembersMutation } from "@/services/group-role/group-role.mutation";

type RoleAddMembersDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	groupId: string;
	roleId: string;
	roleLabel: string;
	allMembers: GroupMember[];
	assignedUserIds: Set<string>;
	hideRoles?: boolean;
};

export function RoleAddMembersDialog({
	open,
	onOpenChange,
	groupId,
	roleId,
	roleLabel,
	allMembers,
	assignedUserIds,
	hideRoles = false,
}: RoleAddMembersDialogProps) {
	const { t } = useTranslation();
	const titleId = useId();
	const searchId = useId();
	const [search, setSearch] = useState("");
	const [selectedIds, setSelectedIds] = useState<string[]>([]);
	const addMutation = useAddGroupRoleMembersMutation(groupId, roleId);

	const candidates = useMemo(() => {
		const q = search.trim().toLowerCase();
		return allMembers.filter((m) => {
			if (assignedUserIds.has(m.user_id)) {
				return false;
			}
			if (!q) {
				return true;
			}
			return (
				m.display_name.toLowerCase().includes(q) || (m.username ?? "").toLowerCase().includes(q)
			);
		});
	}, [allMembers, assignedUserIds, search]);

	const toggleSelect = (userId: string) => {
		setSelectedIds((prev) =>
			prev.includes(userId) ? prev.filter((id) => id !== userId) : [...prev, userId],
		);
	};

	const handleOpenChange = (next: boolean) => {
		if (!next) {
			setSearch("");
			setSelectedIds([]);
		}
		onOpenChange(next);
	};

	const handleSubmit = async () => {
		if (selectedIds.length === 0) {
			return;
		}
		try {
			await addMutation.mutateAsync(selectedIds);
			toast.success(
				t("group_settings.roles.detail.add_members_success", {
					count: selectedIds.length,
					name: roleLabel,
				}),
			);
			handleOpenChange(false);
		} catch (error) {
			toast.error(error instanceof ApiError ? error.message : t("common.error_generic"));
		}
	};

	return (
		<WiChatDialog open={open} onOpenChange={handleOpenChange} titleId={titleId}>
			<WiChatDialogHeader
				titleId={titleId}
				title={t("group_settings.roles.detail.add_members_dialog_title", { name: roleLabel })}
				subtitle={t("group_settings.roles.detail.add_members_dialog_desc")}
			/>

			<WiChatDialogBody className="flex flex-col gap-3">
				<label className="relative block min-w-0" htmlFor={searchId}>
					<Search
						className="pointer-events-none absolute top-1/2 left-3.5 z-[1] size-4 -translate-y-1/2 text-muted-foreground"
						aria-hidden
					/>
					<Input
						id={searchId}
						value={search}
						onChange={(e) => setSearch(e.target.value)}
						placeholder={t("group_settings.roles.detail.add_members_dialog_search")}
						className={cn(authInputClass, groupSettingsToolbarSearchInputClass)}
						autoComplete="off"
					/>
				</label>

				<section className="space-y-2">
					<h3 className={authFieldLabelClass}>{t("invite.section_members")}</h3>
					<div className="thin-scroll flex max-h-64 flex-col divide-y divide-border/80 overflow-y-auto rounded-xl bg-field/40">
						{candidates.length === 0 ? (
							<div className="px-4 py-8 text-center text-sm text-muted-foreground">
								{t("group_settings.roles.detail.add_members_dialog_empty")}
							</div>
						) : (
							candidates.map((member) => {
								const name = member.display_name.trim() || t("invite.member_name_fallback");
								const checked = selectedIds.includes(member.user_id);
								const nameColor = resolveMemberNameColor(member.roles, hideRoles);
								return (
									<button
										key={member.user_id}
										type="button"
										onClick={() => toggleSelect(member.user_id)}
										aria-pressed={checked}
										className="flex items-center justify-between gap-3 px-3 py-2.5 text-left transition-colors first:rounded-t-xl last:rounded-b-xl hover:bg-background/50"
									>
										<div className="flex min-w-0 items-center gap-3">
											<UserAvatar
												src={member.avatar_url}
												initials={userInitials(name, member.username)}
												size="md"
												status={member.presence_status}
												statusRingClassName="ring-card"
											/>
											<div className="min-w-0">
												<div className="flex min-w-0 items-center gap-1.5">
													<p
														className="truncate text-sm font-medium text-foreground"
														style={nameColor ? { color: nameColor } : undefined}
													>
														{name}
													</p>
													{member.role === "owner" ? (
														<OwnerCrownIcon label={t("invite.role_owner")} />
													) : null}
												</div>
												{member.username ? (
													<p className="truncate text-xs text-muted-foreground">
														@{member.username}
													</p>
												) : null}
											</div>
										</div>
										<span
											className={cn(
												"inline-flex size-5 shrink-0 items-center justify-center rounded-md border transition-colors",
												checked
													? "border-primary bg-primary text-primary-foreground"
													: "border-border bg-background",
											)}
											aria-hidden
										>
											{checked ? <Check className="size-3.5" strokeWidth={2.5} /> : null}
										</span>
									</button>
								);
							})
						)}
					</div>
				</section>
			</WiChatDialogBody>

			<WiChatDialogFooter>
				<WiChatDialogFooterActions
					backLabel={t("group_settings.roles.delete_dialog.cancel")}
					onBack={() => handleOpenChange(false)}
					primaryLabel={t("group_settings.roles.detail.add_members_dialog_submit", {
						count: selectedIds.length,
					})}
					primaryType="button"
					primaryDisabled={selectedIds.length === 0 || addMutation.isPending}
					onPrimary={() => void handleSubmit()}
					backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
					primaryClassName={cn(authSubmitClass, authFormFooterButtonClass)}
				/>
			</WiChatDialogFooter>
		</WiChatDialog>
	);
}
