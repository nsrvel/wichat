"use client";

import { Plus, X } from "lucide-react";
import { useTranslation } from "react-i18next";

type UserStatusNotePillProps = {
	label?: string;
	onEdit?: () => void;
	onClear?: () => void;
};

export function UserStatusNotePill({ label, onEdit, onClear }: UserStatusNotePillProps) {
	const { t } = useTranslation();
	const trimmed = label?.trim() || "";
	const content = (
		<>
			{!trimmed ? <Plus className="size-3.5 shrink-0 opacity-80" aria-hidden /> : null}
			<span className="line-clamp-2 break-words leading-snug">
				{trimmed || t("shell.user_dock.status_note_placeholder")}
			</span>
		</>
	);

	return (
		<div className="min-w-0 flex-1">
			<div className="inline-flex w-fit max-w-full items-center gap-1.5 rounded-2xl bg-muted px-3.5 py-2.5 text-xs text-muted-foreground shadow-[0_6px_16px_rgba(0,0,0,0.28)] transition-colors hover:bg-muted/85 hover:text-foreground">
				{onEdit ? (
					<button
						type="button"
						onClick={onEdit}
						className="flex min-w-0 items-center gap-1.5 text-left"
					>
						{content}
					</button>
				) : (
					<div className="flex min-w-0 items-center gap-1.5 text-left">{content}</div>
				)}
				{trimmed && onClear ? (
					<button
						type="button"
						onClick={onClear}
						aria-label={t("shell.user_dock.status_note_clear")}
						className="shrink-0 rounded-full p-0.5 text-muted-foreground/80 transition-colors hover:bg-foreground/10 hover:text-foreground"
					>
						<X className="size-3.5" aria-hidden />
					</button>
				) : null}
			</div>
		</div>
	);
}

