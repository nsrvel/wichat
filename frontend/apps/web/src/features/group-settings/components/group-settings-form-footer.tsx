"use client";

import type { FieldErrors } from "react-hook-form";
import { useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import {
	groupSettingsFormFooterBarClass,
	groupSettingsFormFooterBarNudgeClass,
} from "@/features/group-settings/lib/group-settings-chrome";
import type { GroupSettingsProfileFormValues } from "@/features/group-settings/lib/group-settings-profile-schema";
import { useGroupSettingsUnsaved } from "@/features/group-settings/lib/group-settings-unsaved-guard";
import { translateFormFieldMessage } from "@/integrations/i18n/translate-form-field-message";
import { cn } from "@/lib/utils";

function firstFormErrorKey(
	errors: FieldErrors<GroupSettingsProfileFormValues>,
): string | undefined {
	if (errors.name?.message) {
		return String(errors.name.message);
	}
	if (errors.description?.message) {
		return String(errors.description.message);
	}
	return undefined;
}

type GroupSettingsFormFooterProps = {
	formId: string;
	errors: FieldErrors<GroupSettingsProfileFormValues>;
	onReset: () => void;
	isPending?: boolean;
};

export function GroupSettingsFormFooter({
	formId,
	errors,
	onReset,
	isPending = false,
}: GroupSettingsFormFooterProps) {
	const { t } = useTranslation();
	const { footerNudgeKey } = useGroupSettingsUnsaved();
	const errorMessage = translateFormFieldMessage(t, firstFormErrorKey(errors));
	const barRef = useRef<HTMLDivElement>(null);
	const [prevKey, setPrevKey] = useState(footerNudgeKey);
	const [nudgeActive, setNudgeActive] = useState(false);

	if (footerNudgeKey !== prevKey) {
		setPrevKey(footerNudgeKey);
		if (footerNudgeKey > 0) {
			setNudgeActive(true);
		}
	}

	useEffect(() => {
		if (!nudgeActive) {
			return;
		}
		barRef.current?.scrollIntoView({ block: "nearest", behavior: "smooth" });
		const timer = window.setTimeout(() => setNudgeActive(false), 2800);
		return () => window.clearTimeout(timer);
	}, [nudgeActive]);

	const showUnsavedNudge = nudgeActive;
	const statusMessage = showUnsavedNudge
		? t("group_settings.unsaved_block")
		: errorMessage
			? errorMessage
			: "\u00a0";

	return (
		<div
			ref={barRef}
			key={
				showUnsavedNudge
					? `group-settings-footer-nudge-${footerNudgeKey}`
					: "group-settings-footer-idle"
			}
			className={cn(
				groupSettingsFormFooterBarClass,
				"motion-safe:transition-[box-shadow,border-color,background-color] motion-safe:duration-300",
				showUnsavedNudge && groupSettingsFormFooterBarNudgeClass,
				showUnsavedNudge && "group-settings-footer-nudge",
			)}
		>
			<p
				className={cn(
					"flex min-w-0 flex-1 items-center text-sm leading-snug motion-safe:transition-colors motion-safe:duration-200",
					showUnsavedNudge
						? "font-medium text-foreground"
						: errorMessage
							? "text-destructive"
							: "text-transparent",
				)}
				role={showUnsavedNudge ? "status" : errorMessage ? "alert" : undefined}
				aria-live={showUnsavedNudge || errorMessage ? "polite" : undefined}
			>
				{statusMessage}
			</p>
			<div className="flex shrink-0 items-center gap-2">
				<Button
					type="button"
					variant="ghost"
					className="text-muted-foreground hover:text-foreground"
					onClick={onReset}
					disabled={isPending}
				>
					{t("group_settings.reset")}
				</Button>
				<Button type="submit" form={formId} disabled={isPending}>
					{t("group_settings.save_changes")}
				</Button>
			</div>
		</div>
	);
}
