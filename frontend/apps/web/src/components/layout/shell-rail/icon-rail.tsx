"use client";

import { Plus } from "lucide-react";
import { useState } from "react";
import { WiStudioLogoMark } from "@/components/brand/wistudio-logo-mark";
import {
	IconRailActionRow,
	IconRailWorkspaceRow,
} from "@/components/layout/shell-rail/icon-rail-row";
import { IconRailTooltipTile } from "@/components/layout/shell-rail/icon-rail-tile-button";
import {
	shellRailAsideClass,
	shellRailAsideInnerClass,
	shellRailChromeState,
} from "@/components/layout/shell-rail/shell-rail-aside";
import { SHELL_RAIL_TILE_PX } from "@/components/layout/shell-rail/shell-rail-metrics";
const railLogoTileClass =
	"relative overflow-hidden border-0 bg-transparent p-0 shadow-none hover:opacity-90";

type RailWorkspaceId = "wistudio";

type IconRailProps = {
	open: boolean;
};

export function IconRail({ open }: IconRailProps) {
	const [activeWorkspace, setActiveWorkspace] = useState<RailWorkspaceId>("wistudio");

	return (
		<aside
			className={shellRailAsideClass()}
			data-state={shellRailChromeState(open)}
			aria-label="Workspaces"
			aria-hidden={!open}
		>
			<div className={shellRailAsideInnerClass()}>
				<IconRailWorkspaceRow selected={activeWorkspace === "wistudio"}>
					<IconRailTooltipTile
						variant="tile"
						className={railLogoTileClass}
						label="WiStudio"
						aria-current={activeWorkspace === "wistudio" ? "true" : undefined}
						onClick={() => setActiveWorkspace("wistudio")}
					>
						<WiStudioLogoMark size={SHELL_RAIL_TILE_PX} />
					</IconRailTooltipTile>
				</IconRailWorkspaceRow>
				<IconRailActionRow>
					<IconRailTooltipTile variant="muted" label="Add workspace">
						<Plus aria-hidden className="h-[18px] w-[18px]" strokeWidth={2} />
					</IconRailTooltipTile>
				</IconRailActionRow>
			</div>
		</aside>
	);
}
