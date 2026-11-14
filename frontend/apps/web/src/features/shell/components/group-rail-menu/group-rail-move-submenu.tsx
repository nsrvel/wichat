"use client";

import { ChevronRight } from "lucide-react";
import { useTranslation } from "react-i18next";
import { useGroupMenuKit } from "@/features/shell/components/group-rail-menu/group-menu-kit";
import {
	menuChevronClass,
	menuItemClass,
	menuPopupClass,
	menuPositionerClass,
	submenuTriggerClass,
} from "@/lib/menu-chrome";
import { railOrderMove } from "@/features/shell/lib/rail-order";
import { useReorderMyGroupsMutation } from "@/services/group/group.mutation";

type GroupRailMoveSubmenuProps = {
	groupId: string;
	orderedGroupIds: string[];
	onMoved: () => void;
};

export function GroupRailMoveSubmenu({
	groupId,
	orderedGroupIds,
	onMoved,
}: GroupRailMoveSubmenuProps) {
	const { t } = useTranslation();
	const kit = useGroupMenuKit();
	const reorder = useReorderMyGroupsMutation();

	const move = (where: "first" | "last") => {
		const next = railOrderMove(orderedGroupIds, groupId, where);
		if (next.join() === orderedGroupIds.join()) {
			return;
		}
		reorder.mutate(next, { onSuccess: onMoved });
	};

	return (
		<kit.SubmenuRoot>
			<kit.SubmenuTrigger className={submenuTriggerClass} disabled={reorder.isPending}>
				<span className="min-w-0 flex-1">{t("shell.group_rail_menu.move_to")}</span>
				<ChevronRight className={menuChevronClass} aria-hidden />
			</kit.SubmenuTrigger>
			<kit.Portal>
				<kit.Positioner className={menuPositionerClass} alignOffset={-4} sideOffset={-4}>
					<kit.Popup className={menuPopupClass}>
						<kit.Item
							className={menuItemClass}
							disabled={reorder.isPending || orderedGroupIds[0] === groupId}
							onClick={() => move("first")}
						>
							{t("shell.group_rail_menu.move_first")}
						</kit.Item>
						<kit.Item
							className={menuItemClass}
							disabled={
								reorder.isPending || orderedGroupIds[orderedGroupIds.length - 1] === groupId
							}
							onClick={() => move("last")}
						>
							{t("shell.group_rail_menu.move_last")}
						</kit.Item>
					</kit.Popup>
				</kit.Positioner>
			</kit.Portal>
		</kit.SubmenuRoot>
	);
}
