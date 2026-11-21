"use client";

import { GroupMemberSchema } from "@wichat/contracts/group/member";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { useMemo } from "react";
import type { z } from "zod";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { OwnerCrownIcon } from "@/components/ui/owner-crown-icon";
import { UserAvatar } from "@/components/ui/user-avatar";
import { groupMembersForPanel, resolveMemberNameColor } from "@/lib/member-role-display";
import { memberPresenceDot } from "@/features/presence/presence-dot";
import { userInitials } from "@/lib/user-initials";

type Member = z.infer<typeof GroupMemberSchema>;

type GroupAccessListProps = {
	members: Member[];
	loading: boolean;
	failed?: boolean;
	filterQuery?: string;
	hideRoles?: boolean;
	hasPrev: boolean;
	hasNext: boolean;
	onPrev: () => void;
	onNext: () => void;
	onSelectMember?: (member: Member) => void;
};

function MemberAvatar({
	displayName,
	username,
	avatarUrl,
	member,
	onlineLabel,
}: {
	displayName: string;
	username: string;
	avatarUrl?: string;
	member: Pick<Member, "is_online" | "presence_status">;
	onlineLabel: string;
}) {
	const dot = memberPresenceDot(member);
	return (
		<div className="relative shrink-0" aria-label={member.is_online ? onlineLabel : undefined}>
			<UserAvatar
				src={avatarUrl}
				initials={userInitials(displayName, username)}
				size="md"
				status={dot}
				statusRingClassName="ring-card"
			/>
		</div>
	);
}

const rowSkeletonClass =
	"flex items-center gap-3 px-3 py-2.5 animate-pulse first:rounded-t-xl last:rounded-b-xl";

function compareMembersOwnerFirst(a: Member, b: Member): number {
	const rank = (role: Member["role"]) => (role === "owner" ? 0 : 1);
	const byRole = rank(a.role) - rank(b.role);
	if (byRole !== 0) {
		return byRole;
	}
	return a.display_name.localeCompare(b.display_name, undefined, { sensitivity: "base" });
}

export function GroupAccessList({
	members,
	loading,
	failed = false,
	filterQuery = "",
	hideRoles = false,
	hasPrev,
	hasNext,
	onPrev,
	onNext,
	onSelectMember,
}: GroupAccessListProps) {
	const { t } = useTranslation();
	const filter = filterQuery.trim().toLowerCase();
	const visibleMembers = useMemo(() => {
		const rows = !filter
			? [...members]
			: members.filter((member) => {
					const name = member.display_name.trim().toLowerCase();
					const username = (member.username ?? "").trim().toLowerCase();
					return name.includes(filter) || username.includes(filter);
				});
		rows.sort(compareMembersOwnerFirst);
		return rows;
	}, [filter, members]);
	const filtering = filter.length > 0;
	const showPager = !filtering && (hasPrev || hasNext);
	const showFilterEmpty = filtering && !loading && !failed && visibleMembers.length === 0;
	const groupedSections = useMemo(() => {
		if (hideRoles || filtering || visibleMembers.length === 0) {
			return null;
		}
		const groups = groupMembersForPanel(visibleMembers, hideRoles, {
			online: t("invite.member_online", "Online"),
			offline: t("invite.member_offline", "Offline"),
		});
		const hasHoisted = groups.some((g) => g.kind === "hoisted");
		return hasHoisted ? groups : null;
	}, [filtering, hideRoles, t, visibleMembers]);

	const renderMemberRow = (member: Member) => {
		const nameColor = resolveMemberNameColor(member.roles, hideRoles);
		const rowBody = (
			<>
				<MemberAvatar
					displayName={member.display_name}
					username={member.username}
					avatarUrl={member.avatar_url}
					member={member}
					onlineLabel={t("invite.member_online")}
				/>
				<div className="min-w-0 flex-1">
					<div className="flex min-w-0 items-center gap-1.5">
						<p
							className="truncate text-sm font-medium text-foreground"
							style={nameColor ? { color: nameColor } : undefined}
						>
							{member.display_name.trim() || t("invite.member_name_fallback")}
						</p>
						{member.role === "owner" ? <OwnerCrownIcon label={t("invite.role_owner")} /> : null}
					</div>
					{member.username ? (
						<p
							className="truncate select-text text-xs text-muted-foreground"
							data-wichat-selectable="true"
						>
							@{member.username}
						</p>
					) : null}
				</div>
			</>
		);

		return (
			<li key={member.user_id} className="first:rounded-t-xl last:rounded-b-xl">
				{onSelectMember ? (
					<button
						type="button"
						onClick={() => onSelectMember(member)}
						className="flex w-full cursor-pointer items-center gap-3 px-3 py-2.5 text-left transition-colors hover:bg-background/40"
					>
						{rowBody}
					</button>
				) : (
					<div className="flex items-center gap-3 px-3 py-2.5">{rowBody}</div>
				)}
			</li>
		);
	};

	return (
		<div className="space-y-2">
			{showPager ? (
				<div className="flex justify-end gap-0.5">
					<Button
						type="button"
						variant="ghost"
						size="icon"
						className="size-8 text-muted-foreground"
						disabled={!hasPrev || loading}
						aria-label={t("invite.members_prev")}
						onClick={onPrev}
					>
						<ChevronLeft className="size-4" aria-hidden />
					</Button>
					<Button
						type="button"
						variant="ghost"
						size="icon"
						className="size-8 text-muted-foreground"
						disabled={!hasNext || loading}
						aria-label={t("invite.members_next")}
						onClick={onNext}
					>
						<ChevronRight className="size-4" aria-hidden />
					</Button>
				</div>
			) : null}

			{showFilterEmpty ? (
				<p className="py-8 text-center text-sm text-muted-foreground sm:py-10" role="status">
					{t("invite.members_filter_empty")}
				</p>
			) : (
				<ul
					className="divide-y divide-border/80 rounded-xl bg-field/40"
					aria-busy={loading && !failed}
					aria-label={t("invite.section_members_label")}
				>
					{loading && visibleMembers.length === 0
						? [0, 1, 2].map((key) => (
								<li key={key} className={rowSkeletonClass}>
									<div className="size-9 shrink-0 rounded-full bg-muted" aria-hidden />
									<div className="min-w-0 flex-1">
										<div className="h-3.5 w-28 rounded bg-muted" />
									</div>
								</li>
							))
						: null}
					{groupedSections
						? groupedSections.map((section) => (
								<li key={section.key} className="list-none">
									<div className="bg-muted/35 px-3 py-1.5 text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">
										{section.label} — {section.members.length}
									</div>
									<ul className="divide-y divide-border/80">
										{section.members.map(renderMemberRow)}
									</ul>
								</li>
							))
						: visibleMembers.map(renderMemberRow)}
					{failed ? (
						<li className="px-3 py-8 text-center text-sm text-muted-foreground">
							{t("invite.members_load_error")}
						</li>
					) : null}
					{!loading && !failed && visibleMembers.length === 0 ? (
						<li className="px-3 py-8 text-center text-sm text-muted-foreground">
							{t("invite.members_empty")}
						</li>
					) : null}
				</ul>
			)}
		</div>
	);
}
