"use client";

import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { ChevronDown, Download, FileText } from "lucide-react";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import {
	GroupMemberProfileHost,
	type MemberProfileFallbackUser,
} from "@/components/group/group-member-profile-host";
import { Button } from "@/components/ui/button";
import { CustomSelect, type SelectOption } from "@/components/ui/custom-select";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { GroupSettingsAuditLogRow } from "@/features/group-settings/components/audit-logs/group-settings-audit-log-row";
import {
	groupSettingsMainInnerClass,
	groupSettingsMainScrollClass,
	groupSettingsSectionHeaderClass,
	groupSettingsSectionLeadClass,
	groupSettingsSectionTitleClass,
	groupSettingsToolbarActionButtonClass,
} from "@/features/group-settings/lib/group-settings-chrome";
import type { GroupSettingsTarget } from "@/features/group-settings/lib/group-settings-target";
import { authInputClass } from "@/lib/form-field-styles";
import { cn } from "@/lib/utils";
import { groupAuditLogsInfiniteQueryOpts } from "@/services/group/group.query";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";

const AUDIT_ACTION_KEYS = [
	"group.created",
	"group.profile.updated",
	"group.icon.updated",
	"group.banner.updated",
	"group.settings.hide_roles_updated",
	"group.member.joined",
	"group.member.left",
	"group.member.kicked",
	"group.member.banned",
	"group.member.unbanned",
	"group.member.timed_out",
	"group.member.timeout_removed",
	"group.role.created",
	"group.role.updated",
	"group.role.duplicated",
	"group.role.deleted",
	"group.member.role_added",
	"group.member.role_removed",
	"group.invite.created",
	"group.invite.sent",
] as const;

const EXPORT_RANGE_KEYS = ["today", "7d", "30d", "90d"] as const;

type GroupSettingsAuditLogsSectionProps = {
	target: GroupSettingsTarget;
};

export function GroupSettingsAuditLogsSection({ target }: GroupSettingsAuditLogsSectionProps) {
	const { t } = useTranslation();
	const [actorFilter, setActorFilter] = useState("");
	const [actionFilter, setActionFilter] = useState("");
	const [exportOpen, setExportOpen] = useState(false);
	const [selectedActor, setSelectedActor] = useState<MemberProfileFallbackUser | null>(null);

	const { data: membersData } = useQuery(groupMembersQueryOpts(target.groupId));
	const { data, isLoading, hasNextPage, isFetchingNextPage, fetchNextPage } = useInfiniteQuery(
		groupAuditLogsInfiniteQueryOpts(target.groupId, actorFilter, actionFilter),
	);

	const logs = useMemo(() => data?.pages.flatMap((page) => page.logs) ?? [], [data]);

	const userOptions = useMemo<SelectOption[]>(() => {
		const base: SelectOption[] = [
			{ value: "", label: t("group_settings.audit_logs.filter_user_all") },
		];
		const seen = new Set<string>();
		for (const member of membersData?.members ?? []) {
			if (seen.has(member.user_id)) {
				continue;
			}
			seen.add(member.user_id);
			const label = member.display_name || member.username;
			base.push({ value: member.user_id, label });
		}
		return base;
	}, [membersData?.members, t]);

	const actionOptions = useMemo<SelectOption[]>(() => {
		return [
			{ value: "", label: t("group_settings.audit_logs.filter_action_all") },
			...AUDIT_ACTION_KEYS.map((key) => ({
				value: key,
				label: t(`group_settings.audit_logs.action_options.${key}`),
			})),
		];
	}, [t]);

	const hasActiveFilter = Boolean(actorFilter || actionFilter);

	return (
		<div className={groupSettingsMainScrollClass}>
			<div className={cn(groupSettingsMainInnerClass, "flex flex-col gap-6")}>
				<header className={cn(groupSettingsSectionHeaderClass, "pb-0 sm:pb-0")}>
					<h2 className={groupSettingsSectionTitleClass}>{t("group_settings.audit_logs.title")}</h2>
					<p className={groupSettingsSectionLeadClass}>{t("group_settings.audit_logs.lead")}</p>
				</header>

				<div className="grid grid-cols-1 items-end gap-3 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]">
					<div className="flex min-w-0 flex-col gap-1">
						<span className="text-xs font-medium text-muted-foreground">
							{t("group_settings.audit_logs.filter_user_label")}
						</span>
						<CustomSelect
							options={userOptions}
							value={actorFilter}
							placeholder={t("group_settings.audit_logs.filter_user_all")}
							onChange={setActorFilter}
							triggerClassName={cn(
								authInputClass,
								"flex w-full min-w-0 items-center justify-between gap-2",
							)}
							portaled
							typeahead
						/>
					</div>

					<div className="flex min-w-0 flex-col gap-1">
						<span className="text-xs font-medium text-muted-foreground">
							{t("group_settings.audit_logs.filter_action_label")}
						</span>
						<CustomSelect
							options={actionOptions}
							value={actionFilter}
							placeholder={t("group_settings.audit_logs.filter_action_all")}
							onChange={setActionFilter}
							triggerClassName={cn(
								authInputClass,
								"flex w-full min-w-0 items-center justify-between gap-2",
							)}
							portaled
							typeahead
						/>
					</div>

					<Popover open={exportOpen} onOpenChange={setExportOpen}>
						<PopoverTrigger
							render={
								<Button
									type="button"
									variant="outline"
									className={cn(groupSettingsToolbarActionButtonClass, "gap-2")}
								>
									<Download className="size-4" aria-hidden />
									<span>{t("group_settings.audit_logs.export_button")}</span>
									<ChevronDown className="size-3.5 text-muted-foreground" aria-hidden />
								</Button>
							}
						/>
						<PopoverContent side="bottom" align="end" sideOffset={8} className="w-56 p-1.5">
							<p className="px-2.5 py-1.5 text-[11px] font-medium text-muted-foreground">
								{t("group_settings.audit_logs.export_range_label")}
							</p>
							<div className="flex flex-col gap-0.5">
								{EXPORT_RANGE_KEYS.map((rangeKey) => (
									<button
										key={rangeKey}
										type="button"
										onClick={() => setExportOpen(false)}
										className="flex w-full cursor-pointer items-center rounded-md px-2.5 py-1.5 text-left text-sm text-foreground transition-colors hover:bg-accent"
									>
										{t(`group_settings.audit_logs.export_ranges.${rangeKey}`)}
									</button>
								))}
							</div>
						</PopoverContent>
					</Popover>
				</div>

				{isLoading ? (
					<div className="flex flex-col gap-2.5">
						<div className="h-16 animate-pulse rounded-xl bg-field/40" />
						<div className="h-16 animate-pulse rounded-xl bg-field/40" />
						<div className="h-16 animate-pulse rounded-xl bg-field/40" />
					</div>
				) : logs.length === 0 ? (
					<div className="flex flex-col items-center justify-center gap-2 rounded-xl border border-dashed border-border/60 bg-field/20 px-6 py-12 text-center">
						<FileText className="size-8 text-muted-foreground/70" aria-hidden />
						<p className="text-sm text-muted-foreground">
							{hasActiveFilter
								? t("group_settings.audit_logs.filter_empty")
								: t("group_settings.audit_logs.empty")}
						</p>
					</div>
				) : (
					<div className="flex flex-col gap-2.5">
						{logs.map((entry) => (
							<GroupSettingsAuditLogRow
								key={entry.id}
								log={entry}
								onActorClick={(clickedLog) =>
									setSelectedActor({
										userId: clickedLog.actor_id,
										username: clickedLog.actor_username,
										displayName: clickedLog.actor_display_name,
										avatarUrl: clickedLog.actor_avatar_url,
									})
								}
							/>
						))}

						{hasNextPage ? (
							<div className="flex justify-center pt-2">
								<Button
									type="button"
									variant="outline"
									size="sm"
									disabled={isFetchingNextPage}
									onClick={() => void fetchNextPage()}
								>
									{t("group_settings.audit_logs.load_more")}
								</Button>
							</div>
						) : null}
					</div>
				)}
			</div>

			<GroupMemberProfileHost
				groupId={target.groupId}
				memberId={selectedActor?.userId ?? null}
				fallbackUser={selectedActor}
				onClose={() => setSelectedActor(null)}
			/>
		</div>
	);
}
