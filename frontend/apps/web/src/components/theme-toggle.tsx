"use client";

import { Moon, Sun } from "lucide-react";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import {
	WorkspaceHeaderIconButton,
	workspaceHeaderIconClass,
} from "@/components/layout/workspace-header-icon-button";
import { CustomSelect } from "@/components/ui/custom-select";
import { authSelectGhostTriggerClass } from "@/lib/form-field-styles";
import { useMounted } from "@/hooks/use-mounted";
import { cn } from "@/lib/utils";
import { useTheme } from "next-themes";

type ThemeChoice = "system" | "light" | "dark";

/** Temporary shell control: flip light ↔ dark (ignores system). */
export function ThemeIconToggle() {
	const mounted = useMounted();
	const { resolvedTheme, setTheme } = useTheme();

	if (!mounted) {
		return <span className="inline-block size-9 shrink-0" aria-hidden suppressHydrationWarning />;
	}

	const isDark = resolvedTheme === "dark";

	return (
		<WorkspaceHeaderIconButton
			aria-label={isDark ? "Switch to light mode" : "Switch to dark mode"}
			onClick={() => setTheme(isDark ? "light" : "dark")}
		>
			{isDark ? (
				<Sun className={workspaceHeaderIconClass} strokeWidth={2} />
			) : (
				<Moon className={workspaceHeaderIconClass} strokeWidth={2} />
			)}
		</WorkspaceHeaderIconButton>
	);
}

export function ThemeToggle({
	className,
	triggerClassName,
	compact = true,
}: {
	className?: string;
	triggerClassName?: string;
	compact?: boolean;
}) {
	const { t } = useTranslation();
	const mounted = useMounted();
	const { theme, setTheme } = useTheme();

	const options = useMemo(
		() => [
			{
				value: "system",
				label: t("theme.system"),
				triggerLabel: t("theme.system_short"),
			},
			{
				value: "light",
				label: t("theme.light"),
				triggerLabel: t("theme.light_short"),
			},
			{
				value: "dark",
				label: t("theme.dark"),
				triggerLabel: t("theme.dark_short"),
			},
		],
		[t],
	);

	if (!mounted) {
		return (
			<span
				className={cn("inline-block h-8 w-10 shrink-0", className)}
				aria-hidden
				suppressHydrationWarning
			/>
		);
	}

	const value = (theme ?? "system") as ThemeChoice;

	return (
		<CustomSelect
			className={cn("w-auto shrink-0", className)}
			options={options}
			value={value}
			placeholder={t("theme.label")}
			onChange={(next) => setTheme(next as ThemeChoice)}
			triggerClassName={cn(authSelectGhostTriggerClass, triggerClassName)}
			compactTrigger={compact}
			portaled
			menuAlign="end"
			typeahead
		/>
	);
}
