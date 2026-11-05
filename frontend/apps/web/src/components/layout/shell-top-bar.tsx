"use client";

import { CircleHelp, LogOut, PanelLeft } from "lucide-react";
import { useRouter } from "next/navigation";
import { useTranslation } from "react-i18next";
import { ShellSearchInput } from "@/components/layout/shell-search-input";
import {
	WorkspaceHeaderIconButton,
	workspaceHeaderIconClass,
} from "@/components/layout/workspace-header-icon-button";
import { ThemeIconToggle } from "@/components/theme-toggle";
import { cn } from "@/lib/utils";
import { useLogoutMutation } from "@/services/auth/auth.mutation";

type ShellTopBarProps = {
	railOpen: boolean;
	onRailOpenChange: (open: boolean) => void;
	className?: string;
};

export function ShellTopBar({ railOpen, onRailOpenChange, className }: ShellTopBarProps) {
	const { t } = useTranslation();
	const router = useRouter();
	const logout = useLogoutMutation();

	const signOut = () => {
		logout.mutate(undefined, { onSettled: () => router.replace("/sign-in") });
	};

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
				<ThemeIconToggle />
				<WorkspaceHeaderIconButton
					aria-label={t("logout")}
					disabled={logout.isPending}
					onClick={signOut}
				>
					<LogOut className={workspaceHeaderIconClass} strokeWidth={2} aria-hidden />
				</WorkspaceHeaderIconButton>
				<WorkspaceHeaderIconButton aria-label="Help">
					<CircleHelp className={workspaceHeaderIconClass} strokeWidth={2} />
				</WorkspaceHeaderIconButton>
			</div>
		</header>
	);
}
