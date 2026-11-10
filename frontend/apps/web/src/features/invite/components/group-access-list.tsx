"use client";

import { Menu } from "@base-ui/react/menu";
import { GroupMemberSchema } from "@wichat/contracts";
import { Check, ChevronDown, ChevronLeft, ChevronRight, ChevronUp } from "lucide-react";
import { useMemo, useState } from "react";
import type { z } from "zod";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { userInitials } from "@/features/invite/lib/user-initials";
import {
	menuDropdownTriggerChevronClass,
	menuDropdownTriggerClass,
	menuDropdownTriggerLabelClass,
	menuDropdownTriggerOpenClass,
	menuItemClass,
	menuPopupClass,
	menuPositionerClass,
} from "@/features/shell/components/group-rail-menu/menu-chrome";
import { cn } from "@/lib/utils";

type Member = z.infer<typeof GroupMemberSchema>;

const MEMBER_ROLES = ["owner", "member"] as const;

type GroupAccessListProps = {
	members: Member[];
	loading: boolean;
	failed?: boolean;
	filterQuery?: string;
	canChangeRoles?: boolean;
	hasPrev: boolean;
	hasNext: boolean;
	onPrev: () => void;
	onNext: () => void;
};

function MemberRoleBadge({ role }: { role: Member["role"] }) {
	const { t } = useTranslation();
	const label = t(`invite.role_${role}`);

	return (
		<span
			className="flex min-h-[34px] shrink-0 items-center px-2 py-1.5 lg:min-h-[36px] lg:h-[36px]"
			aria-label={t("invite.member_role_menu", { role: label })}
		>
			<span className={cn(menuDropdownTriggerLabelClass, "max-w-[6.5rem] text-muted-foreground")}>
				{label}
			</span>
		</span>
	);
}

function MemberRoleMenu({ role }: { role: Member["role"] }) {
	const { t } = useTranslation();
	const [open, setOpen] = useState(false);
	const label = t(`invite.role_${role}`);

	return (
		<Menu.Root
			open={open}
			onOpenChange={setOpen}
		>
			<Menu.Trigger
				nativeButton
				render={
					<button
						type="button"
						className={cn(
							menuDropdownTriggerClass,
							"shrink-0",
							open && menuDropdownTriggerOpenClass,
						)}
						aria-haspopup="menu"
						aria-expanded={open}
						aria-label={t("invite.member_role_menu", { role: label })}
					/>
				}
			>
				<span className={cn(menuDropdownTriggerLabelClass, "max-w-[6.5rem]")}>{label}</span>
				{open ? (
					<ChevronUp className={menuDropdownTriggerChevronClass} strokeWidth={2.4} aria-hidden />
				) : (
					<ChevronDown className={menuDropdownTriggerChevronClass} strokeWidth={2.4} aria-hidden />
				)}
			</Menu.Trigger>
			<Menu.Portal>
				<Menu.Positioner className={menuPositionerClass} side="bottom" align="end" sideOffset={4}>
					<Menu.Popup className={menuPopupClass}>
						{MEMBER_ROLES.map((r) => (
							<Menu.Item key={r} className={menuItemClass} closeOnClick={false}>
								<span className="flex min-w-0 flex-1 items-center justify-between gap-2">
									<span>{t(`invite.role_${r}`)}</span>
									{r === role ? (
										<Check className="h-4 w-4 shrink-0 text-primary" aria-hidden />
									) : null}
								</span>
							</Menu.Item>
						))}
					</Menu.Popup>
				</Menu.Positioner>
			</Menu.Portal>
		</Menu.Root>
	);
}

const rowSkeletonClass =
	"flex items-center gap-3 px-3 py-2.5 animate-pulse first:rounded-t-xl last:rounded-b-xl";

export function GroupAccessList({
	members,
	loading,
	failed = false,
	filterQuery = "",
	canChangeRoles = false,
	hasPrev,
	hasNext,
	onPrev,
	onNext,
}: GroupAccessListProps) {
	const { t } = useTranslation();
	const filter = filterQuery.trim().toLowerCase();
	const visibleMembers = useMemo(() => {
		if (!filter) {
			return members;
		}
		return members.filter((member) => {
			const name = member.display_name.trim().toLowerCase();
			const email = (member.email ?? "").trim().toLowerCase();
			return name.includes(filter) || email.includes(filter);
		});
	}, [filter, members]);
	const filtering = filter.length > 0;
	const showPager = !filtering && (hasPrev || hasNext);
	const showFilterEmpty =
		filtering && !loading && !failed && visibleMembers.length === 0;

	return (
		<div className="space-y-2">
			{showPager ? (
				<div className="flex justify-end gap-0.5">
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
									<div className="h-9 w-9 shrink-0 rounded-full bg-muted" aria-hidden />
									<div className="min-w-0 flex-1">
										<div className="h-3.5 w-28 rounded bg-muted" />
									</div>
									<div className="h-5 w-14 shrink-0 rounded-md bg-muted" aria-hidden />
								</li>
							))
						: null}
					{visibleMembers.map((member) => (
						<li
							key={member.user_id}
							className="flex items-center gap-3 px-3 py-2.5 first:rounded-t-xl last:rounded-b-xl"
						>
							<div
								className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-background text-xs font-semibold text-foreground ring-1 ring-border/60"
								aria-hidden
							>
								{userInitials(member.display_name, member.email)}
							</div>
							<div className="min-w-0 flex-1">
								<p className="truncate text-sm font-medium text-foreground">
									{member.display_name.trim() || t("invite.member_name_fallback")}
								</p>
							</div>
							{canChangeRoles ? (
							<MemberRoleMenu role={member.role} />
						) : (
							<MemberRoleBadge role={member.role} />
						)}
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
