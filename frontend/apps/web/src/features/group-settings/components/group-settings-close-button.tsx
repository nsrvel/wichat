"use client";

import { X } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { groupSettingsCloseButtonClass } from "@/features/group-settings/lib/group-settings-chrome";
import { cn } from "@/lib/utils";

type GroupSettingsCloseButtonProps = {
	onClose: () => void;
};

export function GroupSettingsCloseButton({ onClose }: GroupSettingsCloseButtonProps) {
	const { t } = useTranslation();

	return (
		<Button
			type="button"
			variant="ghost"
			className={cn(groupSettingsCloseButtonClass, "active:bg-muted dark:active:bg-muted/50")}
			onClick={onClose}
			aria-label={t("group_settings.close")}
		>
			<span className="text-[10px] leading-none font-semibold tracking-[0.18em] uppercase">
				{t("group_settings.esc")}
			</span>
			<X className="size-5 shrink-0" strokeWidth={2.25} aria-hidden />
		</Button>
	);
}
