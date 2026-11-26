"use client";

import { useTranslation } from "react-i18next";
import {
	isShellSearchFilterEnabled,
	SHELL_SEARCH_FILTERS,
	type ShellSearchFilterId,
} from "@/features/search/lib/shell-search-filters";
import { cn } from "@/lib/utils";

type ShellSearchFilterChipsProps = {
	active: ShellSearchFilterId;
	onChange: (id: ShellSearchFilterId) => void;
};

const FILTER_I18N: Record<ShellSearchFilterId, string> = {
	messages: "shell.search.filter_messages",
	files: "shell.search.filter_files",
	channels: "shell.search.filter_channels",
	images: "shell.search.filter_images",
};

export function ShellSearchFilterChips({ active, onChange }: ShellSearchFilterChipsProps) {
	const { t } = useTranslation();

	return (
		<div
			className="overlay-scroll -mx-0.5 flex gap-2.5 overflow-x-auto px-0.5 pb-0.5"
			role="tablist"
			aria-label={t("shell.search.filters_label")}
		>
			{SHELL_SEARCH_FILTERS.map((id) => {
				const enabled = isShellSearchFilterEnabled(id);
				const selected = active === id;
				return (
					<button
						key={id}
						type="button"
						role="tab"
						aria-selected={selected}
						disabled={!enabled}
						onClick={() => onChange(id)}
						className={cn(
							"shrink-0 rounded-full px-3.5 py-1.5 text-xs font-medium transition-colors",
							selected
								? "bg-primary text-primary-foreground shadow-sm"
								: "bg-muted/80 text-muted-foreground hover:bg-muted hover:text-foreground",
							!enabled &&
								"cursor-not-allowed opacity-50 hover:bg-muted/80 hover:text-muted-foreground",
						)}
					>
						{t(FILTER_I18N[id])}
					</button>
				);
			})}
		</div>
	);
}
