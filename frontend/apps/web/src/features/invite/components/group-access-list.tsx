"use client";

import { GroupMemberSchema } from "@wichat/contracts";
import { ChevronLeft, ChevronRight, Crown } from "lucide-react";
import Image from "next/image";
import { useMemo } from "react";
import type { z } from "zod";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { userInitials } from "@/features/invite/lib/user-initials";
import {
	menuDropdownTriggerClass,
	menuDropdownTriggerLabelClass,
} from "@/features/shell/components/group-rail-menu/menu-chrome";
import { cn } from "@/lib/utils";

type Member = z.infer<typeof GroupMemberSchema>;

type GroupAccessListProps = {
	members: Member[];
	loading: boolean;
	failed?: boolean;
	filterQuery?: string;
	hasPrev: boolean;
	hasNext: boolean;
	onPrev: () => void;
	onNext: () => void;
};

function MemberAvatar({
	displayName,
	username,
	avatarUrl,
	isOnline,
	onlineLabel,
}: {
	displayName: string;
	username: string;
	avatarUrl?: string;
	isOnline?: boolean;
	onlineLabel: string;
}) {
	return (
		<div className="relative shrink-0">
			{avatarUrl ? (
				<Image
					src={avatarUrl}
					alt=""
					width={36}
					height={36}
					unoptimized
					className="size-9 rounded-full object-cover ring-1 ring-border/60"
				/>
			) : (
				<div
					className="flex size-9 items-center justify-center rounded-full bg-background text-xs font-semibold text-foreground ring-1 ring-border/60"
					aria-hidden
				>
					{userInitials(displayName, username)}
				</div>
			)}
			{isOnline ? (
				<span
					className="absolute right-0 bottom-0 size-2.5 rounded-full border-2 border-[color-mix(in_oklab,var(--card)_88%,var(--background))] bg-presence-online"
					role="img"
					aria-label={onlineLabel}
				/>
			) : null}
		</div>
	);
}

function MemberOwnerBadge() {
	const { t } = useTranslation();
	const label = t("invite.role_owner");

	return (
		<span
			className={cn(
				menuDropdownTriggerClass,
				"pointer-events-none shrink-0 cursor-default px-2.5 hover:bg-transparent",
			)}
			aria-label={t("invite.member_role_menu", { role: label })}
		>
			<Crown
				className="size-3.5 shrink-0 text-amber-500 dark:text-amber-400"
				strokeWidth={2.2}
				aria-hidden
			/>
			<span className={cn(menuDropdownTriggerLabelClass, "max-w-[6.5rem]")}>{label}</span>
		</span>
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
	hasPrev,
	hasNext,
	onPrev,
	onNext,
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
					{visibleMembers.map((member) => (
						<li
							key={member.user_id}
							className="flex items-center gap-3 px-3 py-2.5 first:rounded-t-xl last:rounded-b-xl"
						>
							<MemberAvatar
								displayName={member.display_name}
								username={member.username}
								avatarUrl={member.avatar_url}
								isOnline={member.is_online === true}
								onlineLabel={t("invite.member_online")}
							/>
							<div className="min-w-0 flex-1">
								<p className="truncate text-sm font-medium text-foreground">
									{member.display_name.trim() || t("invite.member_name_fallback")}
								</p>
								{member.username ? (
									<p className="truncate text-xs text-muted-foreground">@{member.username}</p>
								) : null}
							</div>
							{member.role === "owner" ? <MemberOwnerBadge /> : null}
						</li>
					))}
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
