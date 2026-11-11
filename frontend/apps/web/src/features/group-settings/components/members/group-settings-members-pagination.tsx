"use client";

import { ChevronLeft, ChevronRight } from "lucide-react";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { groupSettingsDataTablePaginationNavClass } from "@/features/group-settings/lib/group-settings-chrome";
import { cn } from "@/lib/utils";

type GroupSettingsMembersPaginationProps = {
	pageIndex: number;
	pageSize: number;
	pageRowCount: number;
	total: number;
	hasPrev: boolean;
	hasNext: boolean;
	loading?: boolean;
	maxVisitedPage: number;
	onPrev: () => void;
	onNext: () => void;
	onGoToPage: (page: number) => void;
	/** Default `members` — use `roles` on the roles settings table. */
	variant?: "members" | "roles";
};

function buildPageWindow(current: number, totalPages: number): number[] {
	if (totalPages <= 7) {
		return Array.from({ length: totalPages }, (_, i) => i + 1);
	}
	const start = Math.max(1, Math.min(current - 2, totalPages - 4));
	const end = Math.min(totalPages, start + 4);
	return Array.from({ length: end - start + 1 }, (_, i) => start + i);
}

export function GroupSettingsMembersPagination({
	pageIndex,
	pageSize,
	pageRowCount,
	total,
	hasPrev,
	hasNext,
	loading = false,
	maxVisitedPage,
	onPrev,
	onNext,
	onGoToPage,
	variant = "members",
}: GroupSettingsMembersPaginationProps) {
	const { t } = useTranslation();
	const paginationLabelKey =
		variant === "roles"
			? "group_settings.roles.pagination_label"
			: "group_settings.members.pagination_label";
	const paginationShowingKey =
		variant === "roles"
			? "group_settings.roles.pagination_showing"
			: "group_settings.members.pagination_showing";

	const totalPages = Math.max(1, Math.ceil(total / pageSize));
	const showingCount = pageRowCount;

	const pageWindow = useMemo(() => buildPageWindow(pageIndex, totalPages), [pageIndex, totalPages]);

	const showNav = total > 0 || hasPrev || hasNext;

	if (!showNav) {
		return null;
	}

	const canGoToPage = (page: number) => {
		if (page === pageIndex || loading) {
			return false;
		}
		if (page <= maxVisitedPage) {
			return true;
		}
		return page === maxVisitedPage + 1 && hasNext;
	};

	return (
		<nav className={groupSettingsDataTablePaginationNavClass} aria-label={t(paginationLabelKey)}>
			<p className="text-sm text-muted-foreground">
				{t(paginationShowingKey, {
					count: showingCount,
					total,
				})}
			</p>
			<div className="flex items-center justify-end gap-0.5">
				<Button
					type="button"
					variant="ghost"
					size="icon"
					className="h-8 w-8 text-muted-foreground"
					disabled={!hasPrev || loading}
					aria-label={t("invite.members_prev")}
					onClick={onPrev}
				>
					<ChevronLeft className="h-4 w-4" aria-hidden />
				</Button>

				<div
					className="flex items-center gap-0.5 px-0.5"
					role="group"
					aria-label={t("invite.members_page", { page: pageIndex })}
				>
					{pageWindow.map((page) => {
						const selected = page === pageIndex;
						const enabled = canGoToPage(page) || selected;
						return (
							<button
								key={page}
								type="button"
								disabled={!enabled}
								aria-current={selected ? "page" : undefined}
								className={cn(
									"flex h-8 min-w-8 items-center justify-center rounded-md px-1.5 text-sm font-medium tabular-nums transition-colors",
									selected
										? "bg-sidebar-accent font-semibold text-sidebar-foreground"
										: "text-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-foreground disabled:pointer-events-none disabled:opacity-40",
								)}
								onClick={() => {
									if (page === pageIndex) {
										return;
									}
									if (page <= maxVisitedPage) {
										onGoToPage(page);
										return;
									}
									if (page === pageIndex + 1 && hasNext) {
										onNext();
									}
								}}
							>
								{page}
							</button>
						);
					})}
				</div>

				<Button
					type="button"
					variant="ghost"
					size="icon"
					className="h-8 w-8 text-muted-foreground"
					disabled={!hasNext || loading}
					aria-label={t("invite.members_next")}
					onClick={onNext}
				>
					<ChevronRight className="h-4 w-4" aria-hidden />
				</Button>
			</div>
		</nav>
	);
}
