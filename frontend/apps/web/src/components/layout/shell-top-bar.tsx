"use client";

import { CircleHelp, PanelLeft } from "lucide-react";
import { useTranslation } from "react-i18next";
import { ShellSearchInput } from "@/components/layout/shell-search-input";
import {
	WorkspaceHeaderIconButton,
	workspaceHeaderIconClass,
} from "@/components/layout/workspace-header-icon-button";
import { cn } from "@/lib/utils";

type ShellTopBarProps = {
	railOpen: boolean;
	onRailOpenChange: (open: boolean) => void;
	className?: string;
};

export function ShellTopBar({ railOpen, onRailOpenChange, className }: ShellTopBarProps) {
	const { t } = useTranslation();

	return (
		<header
			className={cn(
				"grid shrink-0 grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-2 bg-transparent px-3 py-1.5 text-sidebar-foreground md:grid-cols-[1fr_minmax(0,30rem)_1fr] md:gap-3 md:px-4",
				className,
			)}
		>
			<div className="flex items-center gap-2">
				<WorkspaceHeaderIconButton
					aria-label={railOpen ? t("shell.hide_rail") : t("shell.show_rail")}
					aria-pressed={railOpen}
					aria-expanded={railOpen}
					onClick={() => onRailOpenChange(!railOpen)}
				>
					<PanelLeft className={workspaceHeaderIconClass} strokeWidth={2} />
				</WorkspaceHeaderIconButton>
			</div>

			<ShellSearchInput className="hidden min-w-0 justify-self-stretch sm:block md:justify-self-center" />

			<div className="flex items-center justify-end gap-0.5">
				<WorkspaceHeaderIconButton
					aria-label="Help"
					onClick={() =>
						window.open(
							"https://www.linkedin.com/in/putra-rama-setiawan/",
							"_blank",
							"noopener,noreferrer",
						)
					}
				>
					<CircleHelp className={workspaceHeaderIconClass} strokeWidth={2} />
				</WorkspaceHeaderIconButton>
			</div>
		</header>
	);
}
