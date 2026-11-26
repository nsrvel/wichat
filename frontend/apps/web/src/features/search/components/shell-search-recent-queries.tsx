"use client";

import { History, X } from "lucide-react";
import { useTranslation } from "react-i18next";

type ShellSearchRecentQueriesProps = {
	items: string[];
	onSelect: (query: string) => void;
	onClear: () => void;
};

export function ShellSearchRecentQueries({
	items,
	onSelect,
	onClear,
}: ShellSearchRecentQueriesProps) {
	const { t } = useTranslation();

	if (items.length === 0) {
		return (
			<div className="space-y-3">
				<p className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground/90">
					{t("shell.search.recent_section")}
				</p>
				<p className="text-sm text-muted-foreground/85">{t("shell.search.no_recent")}</p>
			</div>
		);
	}

	return (
		<div className="space-y-3">
			<div className="flex items-center justify-between gap-3">
				<p className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground/90">
					{t("shell.search.recent_section")}
				</p>
				<button
					type="button"
					onClick={onClear}
					className="inline-flex items-center gap-1 text-[11px] font-medium text-muted-foreground transition-colors hover:text-foreground"
				>
					<X className="size-3" aria-hidden />
					{t("shell.search.clear_recent")}
				</button>
			</div>
			<ul className="space-y-1">
				{items.map((item) => (
					<li key={item}>
						<button
							type="button"
							onClick={() => onSelect(item)}
							className="flex w-full items-center gap-3 rounded-lg px-2.5 py-2.5 text-left text-sm text-foreground transition-colors hover:bg-muted/70"
						>
							<History className="size-4 shrink-0 text-muted-foreground" aria-hidden />
							<span className="min-w-0 truncate">{item}</span>
						</button>
					</li>
				))}
			</ul>
		</div>
	);
}
