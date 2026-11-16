"use client";

import { useQuery } from "@tanstack/react-query";
import { Search } from "lucide-react";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import type { GroupBanEntry } from "@wichat/contracts/group/member";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { Input } from "@/components/ui/input";
import { GroupSettingsBansTable } from "@/features/group-settings/components/bans/group-settings-bans-table";
import {
	groupSettingsMainInnerClass,
	groupSettingsMainScrollClass,
	groupSettingsSectionHeaderClass,
	groupSettingsSectionLeadClass,
	groupSettingsSectionTitleClass,
	groupSettingsToolbarSearchInputClass,
} from "@/features/group-settings/lib/group-settings-chrome";
import type { GroupSettingsTarget } from "@/features/group-settings/lib/group-settings-target";
import { translateError } from "@/integrations/i18n/translate-error";
import { authInputClass } from "@/lib/form-field-styles";
import { cn } from "@/lib/utils";
import { useUnbanGroupMemberMutation } from "@/services/group/group.mutation";
import { groupBansQueryOpts } from "@/services/group/group.query";

type GroupSettingsBansSectionProps = {
	target: GroupSettingsTarget;
};

export function GroupSettingsBansSection({ target }: GroupSettingsBansSectionProps) {
	const { t } = useTranslation();
	const [searchInput, setSearchInput] = useState("");
	const [debouncedSearch, setDebouncedSearch] = useState("");
	const [pageIndex, setPageIndex] = useState(1);
	const [cursorsByPage, setCursorsByPage] = useState<Record<number, string>>({ 1: "" });
	const [maxVisitedPage, setMaxVisitedPage] = useState(1);
	const [unbanCandidate, setUnbanCandidate] = useState<GroupBanEntry | null>(null);

	useEffect(() => {
		const timer = setTimeout(() => {
			setDebouncedSearch(searchInput.trim());
			setPageIndex(1);
			setCursorsByPage({ 1: "" });
			setMaxVisitedPage(1);
		}, 250);
		return () => clearTimeout(timer);
	}, [searchInput]);

	const activeCursor = cursorsByPage[pageIndex] ?? "";
	const { data, isLoading } = useQuery(
		groupBansQueryOpts(target.groupId, debouncedSearch, activeCursor),
	);
	const unbanMutation = useUnbanGroupMemberMutation(target.groupId);

	const bans = data?.bans ?? [];
	const total = data?.total ?? 0;
	const nextCursor = data?.next_cursor ?? "";
	const hasNext = Boolean(nextCursor);
	const hasPrev = pageIndex > 1;

	const handleGoToPage = (nextPage: number) => {
		if (nextPage === pageIndex) {
			return;
		}
		if (nextPage === maxVisitedPage + 1 && nextCursor) {
			setCursorsByPage((prev) => ({ ...prev, [nextPage]: nextCursor }));
			setMaxVisitedPage(nextPage);
			setPageIndex(nextPage);
			return;
		}
		if (nextPage >= 1 && nextPage <= maxVisitedPage && cursorsByPage[nextPage] !== undefined) {
			setPageIndex(nextPage);
		}
	};

	const handleConfirmUnban = () => {
		if (!unbanCandidate) {
			return;
		}
		const candidate = unbanCandidate;
		const displayName = candidate.display_name || candidate.username;
		unbanMutation.mutate(candidate.user_id, {
			onSuccess: () => {
				setUnbanCandidate(null);
				toast.success(t("group_settings.bans.unban_success", { name: displayName }));
			},
			onError: (error) => {
				toast.error(translateError(t, error));
			},
		});
	};

	const candidateName = unbanCandidate
		? unbanCandidate.display_name || unbanCandidate.username
		: "";
	const candidateUsername = unbanCandidate?.username || candidateName;

	return (
		<div className={groupSettingsMainScrollClass}>
			<div className={cn(groupSettingsMainInnerClass, "flex flex-col gap-6")}>
				<header className={groupSettingsSectionHeaderClass}>
					<h2 className={groupSettingsSectionTitleClass}>{t("group_settings.bans.title")}</h2>
					<p className={groupSettingsSectionLeadClass}>{t("group_settings.bans.lead")}</p>
				</header>

				<div className="relative">
					<Search
						className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-muted-foreground"
						aria-hidden
					/>
					<Input
						type="search"
						value={searchInput}
						onChange={(event) => setSearchInput(event.target.value)}
						placeholder={t("group_settings.bans.search_placeholder")}
						aria-label={t("group_settings.bans.search_placeholder")}
						className={cn(authInputClass, groupSettingsToolbarSearchInputClass)}
					/>
				</div>

				<GroupSettingsBansTable
					bans={bans}
					loading={isLoading}
					total={total}
					pageIndex={pageIndex}
					hasPrev={hasPrev}
					hasNext={hasNext}
					maxVisitedPage={maxVisitedPage}
					searchQuery={debouncedSearch}
					unbanningUserId={unbanMutation.isPending ? unbanCandidate?.user_id : undefined}
					onPrev={() => handleGoToPage(pageIndex - 1)}
					onNext={() => handleGoToPage(pageIndex + 1)}
					onGoToPage={handleGoToPage}
					onRequestUnban={setUnbanCandidate}
				/>
			</div>

			<ConfirmDialog
				open={Boolean(unbanCandidate)}
				onOpenChange={(open) => {
					if (!open && !unbanMutation.isPending) {
						setUnbanCandidate(null);
					}
				}}
				titleId="unban-member-dialog-title"
				title={t("group_settings.bans.unban_dialog.title", { name: candidateName })}
				message={t("group_settings.bans.unban_dialog.description", {
					username: candidateUsername,
				})}
				cancelLabel={t("group_settings.bans.unban_dialog.cancel")}
				confirmLabel={t("group_settings.bans.unban_dialog.confirm")}
				onConfirm={handleConfirmUnban}
				confirmDisabled={unbanMutation.isPending}
				destructive
			/>
		</div>
	);
}
