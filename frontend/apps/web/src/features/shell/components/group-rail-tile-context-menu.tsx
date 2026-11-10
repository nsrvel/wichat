"use client";

import { ContextMenu } from "@base-ui/react/context-menu";
import type { Group } from "@wichat/contracts";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useRef, useState, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { translateError } from "@/integrations/i18n/translate-error";
import { APP_HOME_PATH, channelPath } from "@/lib/app-routes";
import { clearLastChannelIfGroup } from "@/lib/last-channel-storage";
import { cn } from "@/lib/utils";
import {
	useDeleteGroupMutation,
	useLeaveGroupMutation,
} from "@/services/group/group.mutation";
import { myGroupsQueryOpts } from "@/services/group/group.query";

const menuPopupClass =
	"z-50 min-w-[11rem] rounded-lg border border-border bg-popover p-1 text-sm text-popover-foreground shadow-md outline-none";

const menuItemClass =
	"flex w-full cursor-default select-none items-center rounded-md px-2.5 py-2 outline-none data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground";

type GroupRailTileContextMenuProps = {
	group: Group;
	activeGroupId?: string;
	children: ReactNode;
};

export function GroupRailTileContextMenu({
	group,
	activeGroupId,
	children,
}: GroupRailTileContextMenuProps) {
	const { t } = useTranslation();
	const router = useRouter();
	const queryClient = useQueryClient();
	const leaveGroup = useLeaveGroupMutation();
	const deleteGroup = useDeleteGroupMutation();
	const menuActionsRef = useRef<ContextMenu.Root.Actions>(null);
	const [confirmDestructive, setConfirmDestructive] = useState(false);

	// Owners delete; members leave. Missing role (stale API) → delete — backend rejects non-owners.
	const isOwner = group.role !== "member";
	const destructivePending = isOwner ? deleteGroup.isPending : leaveGroup.isPending;

	const afterRemovedFromGroup = async () => {
		clearLastChannelIfGroup(group.id);
		const remaining = await queryClient.fetchQuery(myGroupsQueryOpts());
		if (activeGroupId === group.id) {
			const next = remaining[0];
			if (next) {
				router.push(channelPath(next.id, next.default_channel_id));
			} else {
				router.push(APP_HOME_PATH);
			}
		}
	};

	const runDestructive = () => {
		const mutation = isOwner ? deleteGroup : leaveGroup;
		mutation.mutate(group.id, {
			onSuccess: () => {
				setConfirmDestructive(false);
				menuActionsRef.current?.close();
				void afterRemovedFromGroup();
			},
			onError: (error) => {
				toast.error(translateError(t, error));
			},
		});
	};

	return (
		<ContextMenu.Root
			actionsRef={menuActionsRef}
			onOpenChange={(open) => {
				if (!open) {
					setConfirmDestructive(false);
				}
			}}
		>
			<ContextMenu.Trigger className="flex items-center justify-center">
				{children}
			</ContextMenu.Trigger>
			<ContextMenu.Portal>
				<ContextMenu.Positioner>
					<ContextMenu.Popup className={menuPopupClass}>
						{confirmDestructive ? (
							<>
								<ContextMenu.Group>
									<ContextMenu.GroupLabel className="px-2.5 py-1.5 text-xs font-normal text-muted-foreground">
										{isOwner
											? t("shell.delete_group_confirm", { name: group.name })
											: t("shell.leave_group_confirm", { name: group.name })}
									</ContextMenu.GroupLabel>
								</ContextMenu.Group>
								<ContextMenu.Item
									className={cn(
										menuItemClass,
										"text-destructive data-[highlighted]:bg-destructive/10",
									)}
									closeOnClick={false}
									disabled={destructivePending}
									onClick={() => runDestructive()}
								>
									{isOwner
										? t("shell.delete_group_action")
										: t("shell.leave_group_action")}
								</ContextMenu.Item>
								<ContextMenu.Item
									className={menuItemClass}
									disabled={destructivePending}
									closeOnClick={false}
									onClick={() => setConfirmDestructive(false)}
								>
									{isOwner
										? t("shell.delete_group_cancel")
										: t("shell.leave_group_cancel")}
								</ContextMenu.Item>
							</>
						) : (
							<ContextMenu.Item
								className={cn(
									menuItemClass,
									"text-destructive data-[highlighted]:bg-destructive/10",
								)}
								closeOnClick={false}
								onClick={() => setConfirmDestructive(true)}
							>
								{isOwner ? t("shell.delete_group") : t("shell.leave_group")}
							</ContextMenu.Item>
						)}
					</ContextMenu.Popup>
				</ContextMenu.Positioner>
			</ContextMenu.Portal>
		</ContextMenu.Root>
	);
}
