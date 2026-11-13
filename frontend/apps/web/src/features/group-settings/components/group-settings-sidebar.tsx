"use client";

import { useQuery } from "@tanstack/react-query";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { toast } from "sonner";
import {
	Ban,
	Compass,
	FileText,
	Shield,
	SquareArrowOutUpRight,
	Trash2,
	User,
	Users,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { useTranslation } from "react-i18next";
import { myGroupsQueryOpts } from "@/services/group/group.query";
import { useDeleteGroupMutation } from "@/services/group/group.mutation";
import { useGroupSettings } from "@/features/group-settings/group-settings-provider";
import { DeleteGroupConfirmDialog } from "@/features/group-settings/components/delete-group-confirm-dialog";
import { translateError } from "@/integrations/i18n/translate-error";
import { APP_HOME_PATH, channelPath } from "@/lib/app-routes";
import { clearLastChannelIfGroup } from "@/lib/last-channel-storage";
import { SidebarPlainNavRow } from "@/components/layout/channel-sidebar/sidebar-plain-nav-row";
import {
	SIDEBAR_INSET_X,
	sidebarRowInnerButtonClass,
	sidebarRowShellClass,
	sidebarRowTextClass,
} from "@/components/layout/channel-sidebar/sidebar-nav-styles";
import { SidebarStaticCategoryRow } from "@/components/layout/channel-sidebar/sidebar-static-category-row";
import { ChannelGlyphSlot, sidebarNavGlyphIconClass } from "@/components/layout/sidebar-list-icon";
import {
	groupSettingsNavGroups,
	type GroupSettingsNavItem,
} from "@/features/group-settings/lib/group-settings-nav";
import type {
	GroupSettingsSection,
	GroupSettingsTarget,
} from "@/features/group-settings/lib/group-settings-target";
import { cn } from "@/lib/utils";

const NAV_ICON: Record<GroupSettingsSection, LucideIcon> = {
	profile: User,
	onboarding: Compass,
	members: Users,
	roles: Shield,
	audit_logs: FileText,
	bans: Ban,
};

type GroupSettingsSidebarProps = {
	target: GroupSettingsTarget;
	activeSection: GroupSettingsSection;
	onSectionChange: (section: GroupSettingsSection) => void;
	onOpenMembersInWorkspace: () => void;
};

function SettingsNavItem({
	item,
	activeSection,
	onSectionChange,
	onOpenMembersInWorkspace,
}: {
	item: GroupSettingsNavItem;
	activeSection: GroupSettingsSection;
	onSectionChange: (section: GroupSettingsSection) => void;
	onOpenMembersInWorkspace: () => void;
}) {
	const { t } = useTranslation();
	const active = item.section === activeSection;
	const Icon = NAV_ICON[item.id];

	return (
		<SidebarPlainNavRow
			icon={<Icon className={sidebarNavGlyphIconClass} strokeWidth={2} />}
			label={t(item.sectionKey)}
			selected={active}
			onSelect={() => {
				if (item.id === "members") {
					onOpenMembersInWorkspace();
				} else {
					onSectionChange(item.section);
				}
			}}
			trailingIcon={
				item.id === "members" ? (
					<SquareArrowOutUpRight className="size-3.5" strokeWidth={2} aria-hidden />
				) : undefined
			}
		/>
	);
}

export function GroupSettingsSidebar({
	target,
	activeSection,
	onSectionChange,
	onOpenMembersInWorkspace,
}: GroupSettingsSidebarProps) {
	const { t } = useTranslation();
	const router = useRouter();
	const queryClient = useQueryClient();
	const { closeGroupSettings } = useGroupSettings();
	const deleteGroupMutation = useDeleteGroupMutation();
	const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);

	const { data: myGroups = [] } = useQuery(myGroupsQueryOpts());
	const isOwner =
		target.role === "owner" || myGroups.some((g) => g.id === target.groupId && g.role === "owner");

	const handleDeleteConfirm = () => {
		deleteGroupMutation.mutate(target.groupId, {
			onSuccess: async () => {
				setDeleteDialogOpen(false);
				closeGroupSettings();
				clearLastChannelIfGroup(target.groupId);
				const remaining = await queryClient.fetchQuery(myGroupsQueryOpts());
				const next = remaining.find((g) => g.id !== target.groupId);
				if (next) {
					router.push(channelPath(next.id, next.default_channel_id));
				} else {
					router.push(APP_HOME_PATH);
				}
			},
			onError: (error) => {
				toast.error(translateError(t, error));
			},
		});
	};

	return (
		<>
			{isOwner ? (
				<DeleteGroupConfirmDialog
					groupName={target.groupName}
					open={deleteDialogOpen}
					onOpenChange={setDeleteDialogOpen}
					onConfirm={handleDeleteConfirm}
					confirmDisabled={deleteGroupMutation.isPending}
				/>
			) : null}
			<aside className="flex h-full min-size-full max-h-[100dvh] flex-col">
				<div className={cn("flex min-h-0 flex-1 flex-col pt-12 sm:pt-14", SIDEBAR_INSET_X)}>
					<div className="thin-scroll flex min-h-0 flex-1 flex-col overflow-y-auto pb-3">
						<p className="mb-2 truncate px-0.5 text-sm font-semibold text-sidebar-foreground">
							{target.groupName}
						</p>
						<nav className="flex flex-col" aria-label={t("group_settings.nav_label")}>
							{groupSettingsNavGroups.map((group, index) => (
								<div
									key={group.labelKey}
									className={cn(
										"group/cat",
										index > 0 && "mt-3 border-t border-sidebar-border/35 pt-3",
									)}
								>
									<SidebarStaticCategoryRow title={t(group.labelKey)} />
									<div className="flex flex-col gap-0">
										{group.items
											.filter((item) => (item.id !== "members" && item.id !== "roles") || isOwner)
											.map((item) => (
												<SettingsNavItem
													key={item.id}
													item={item}
													activeSection={activeSection}
													onSectionChange={onSectionChange}
													onOpenMembersInWorkspace={onOpenMembersInWorkspace}
												/>
											))}
									</div>
								</div>
							))}

							{isOwner ? (
								<div className="mt-3 border-t border-sidebar-border/35 pt-3">
									<div
										className={cn(
											sidebarRowShellClass,
											"text-destructive hover:bg-destructive/10 dark:hover:bg-destructive/15",
										)}
									>
										<button
											type="button"
											className={cn(sidebarRowInnerButtonClass, "text-destructive")}
											onClick={() => setDeleteDialogOpen(true)}
										>
											<ChannelGlyphSlot framed={false} className="text-destructive">
												<Trash2 className={sidebarNavGlyphIconClass} strokeWidth={2} aria-hidden />
											</ChannelGlyphSlot>
											<span
												className={cn(
													sidebarRowTextClass,
													"min-w-0 flex-1 truncate text-left font-medium capitalize",
												)}
											>
												{t("shell.delete_group")}
											</span>
										</button>
									</div>
								</div>
							) : null}
						</nav>
					</div>
				</div>
			</aside>
		</>
	);
}
