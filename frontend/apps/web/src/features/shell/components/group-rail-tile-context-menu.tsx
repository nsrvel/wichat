"use client";

import { ContextMenu } from "@base-ui/react/context-menu";
import type { Group } from "@wichat/contracts";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useState, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { APP_HOME_PATH, channelPath } from "@/lib/app-routes";
import { clearLastChannelIfGroup } from "@/lib/last-channel-storage";
import { cn } from "@/lib/utils";
import { useLeaveGroupMutation } from "@/services/group/group.mutation";
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
	const [confirmLeave, setConfirmLeave] = useState(false);

	const runLeave = () => {
		leaveGroup.mutate(group.id, {
			onSuccess: async () => {
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
			},
		});
	};

	return (
		<ContextMenu.Root
			onOpenChange={(open) => {
				if (!open) {
					setConfirmLeave(false);
				}
			}}
		>
			<ContextMenu.Trigger className="flex items-center justify-center">
				{children}
			</ContextMenu.Trigger>
			<ContextMenu.Portal>
				<ContextMenu.Positioner>
					<ContextMenu.Popup className={menuPopupClass}>
						{confirmLeave ? (
							<div className="px-2 py-1.5">
								<p className="mb-2 text-xs text-muted-foreground">
									{t("shell.leave_group_confirm", { name: group.name })}
								</p>
								<div className="flex flex-col gap-0.5">
									<ContextMenu.Item
										className={cn(
											menuItemClass,
											"text-destructive data-[highlighted]:bg-destructive/10",
										)}
										disabled={leaveGroup.isPending}
										onClick={() => runLeave()}
									>
										{t("shell.leave_group_action")}
									</ContextMenu.Item>
									<ContextMenu.Item
										className={menuItemClass}
										disabled={leaveGroup.isPending}
										onClick={() => setConfirmLeave(false)}
										closeOnClick={false}
									>
										{t("shell.leave_group_cancel")}
									</ContextMenu.Item>
								</div>
							</div>
						) : (
							<ContextMenu.Item
								className={cn(
									menuItemClass,
									"text-destructive data-[highlighted]:bg-destructive/10",
								)}
								closeOnClick={false}
								onClick={() => setConfirmLeave(true)}
							>
								{t("shell.leave_group")}
							</ContextMenu.Item>
						)}
					</ContextMenu.Popup>
				</ContextMenu.Positioner>
			</ContextMenu.Portal>
		</ContextMenu.Root>
	);
}
