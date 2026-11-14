"use client";

import { ContextMenu } from "@base-ui/react/context-menu";
import { ChevronRight } from "lucide-react";
import { useTranslation } from "react-i18next";
import {
	menuChevronClass,
	menuHintClass,
	menuItemClass,
	menuPopupClass,
	menuPositionerClass,
	submenuTriggerClass,
} from "@/lib/menu-chrome";

type GroupRailStubSubmenuProps = {
	label: string;
	sublabel?: string;
};

export function GroupRailStubSubmenu({ label, sublabel }: GroupRailStubSubmenuProps) {
	const { t } = useTranslation();

	return (
		<ContextMenu.SubmenuRoot>
			<ContextMenu.SubmenuTrigger className={submenuTriggerClass}>
				{sublabel ? (
					<div className="flex min-w-0 flex-1 flex-col items-start gap-1 py-0.5">
						<span className="leading-tight">{label}</span>
						<span className={menuHintClass}>{sublabel}</span>
					</div>
				) : (
					<span className="min-w-0 flex-1">{label}</span>
				)}
				<ChevronRight className={menuChevronClass} aria-hidden />
			</ContextMenu.SubmenuTrigger>
			<ContextMenu.Portal>
				<ContextMenu.Positioner className={menuPositionerClass} alignOffset={-4} sideOffset={-4}>
					<ContextMenu.Popup className={menuPopupClass}>
						<ContextMenu.Item className={menuItemClass} disabled>
							{t("shell.group_rail_menu.submenu_placeholder")}
						</ContextMenu.Item>
					</ContextMenu.Popup>
				</ContextMenu.Positioner>
			</ContextMenu.Portal>
		</ContextMenu.SubmenuRoot>
	);
}
