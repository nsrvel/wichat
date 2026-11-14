"use client";

import type { GroupMember } from "@wichat/contracts";
import { Search, X } from "lucide-react";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { UserAvatar } from "@/components/ui/user-avatar";
import { authInputClass } from "@/lib/form-field-styles";
import type { GroupSettingsRoleRow } from "@/features/group-settings/lib/group-settings-roles-view";
import { userInitials } from "@/lib/user-initials";
import { cn } from "@/lib/utils";

type RoleDetailMembersTabProps = {
	role: GroupSettingsRoleRow;
	members: GroupMember[];
	canManage: boolean;
};

export function RoleDetailMembersTab({ role, members, canManage }: RoleDetailMembersTabProps) {
	const { t } = useTranslation();
	const [search, setSearch] = useState("");

	const isEveryone = role.id === "everyone";

	const roleMembers = useMemo(() => {
		const byRole = isEveryone ? members : members.filter((m) => m.role === "member");
		const q = search.trim().toLowerCase();
		if (!q) {
			return byRole;
		}
		return byRole.filter(
			(m) =>
				m.display_name.toLowerCase().includes(q) ||
				(m.username ?? "").toLowerCase().includes(q),
		);
	}, [isEveryone, members, search]);

	const handleAddMembers = () => {
		toast.message(t("group_settings.roles.detail.add_members_coming_soon"));
	};

	return (
		<div className="flex flex-col gap-5 pt-5">
			{isEveryone ? (
				<div className="rounded-xl border border-border/80 bg-field/40 px-4 py-3 text-xs leading-relaxed text-muted-foreground">
					{t("group_settings.roles.detail.everyone_members_notice")}
				</div>
			) : null}

			<div className="flex flex-col gap-2.5 sm:flex-row sm:items-center">
				<div className="relative min-w-0 flex-1">
					<Search
						className="pointer-events-none absolute top-1/2 left-3.5 z-[1] size-4 -translate-y-1/2 text-muted-foreground"
						strokeWidth={2}
						aria-hidden
					/>
					<Input
						type="search"
						value={search}
						onChange={(e) => setSearch(e.target.value)}
						placeholder={t("group_settings.roles.detail.members_search_placeholder")}
						className={cn(authInputClass, "pl-9")}
					/>
				</div>

				{!isEveryone && canManage ? (
					<Button type="button" variant="default" className="shrink-0" onClick={handleAddMembers}>
						{t("group_settings.roles.detail.add_members")}
					</Button>
				) : null}
			</div>

			{roleMembers.length === 0 ? (
				<div className="rounded-xl border border-dashed border-border/80 bg-field/30 px-4 py-8 text-center text-sm text-muted-foreground">
					{t("group_settings.roles.detail.members_empty")}
				</div>
			) : (
				<div className="flex flex-col divide-y divide-border/60 overflow-hidden rounded-xl border border-border/80 bg-field/30">
					{roleMembers.map((member) => {
						const name = member.display_name.trim() || t("invite.member_name_fallback");
						return (
							<div
								key={member.user_id}
								className="flex items-center justify-between gap-3 px-4 py-3 transition-colors hover:bg-background/40"
							>
								<div className="flex min-w-0 items-center gap-3">
									<UserAvatar
										src={member.avatar_url}
										initials={userInitials(name, member.username)}
										size="md"
									/>
									<div className="min-w-0">
										<p className="truncate text-sm font-medium text-foreground">{name}</p>
										{member.username ? (
											<p className="truncate text-xs text-muted-foreground">@{member.username}</p>
										) : null}
									</div>
								</div>

								{!isEveryone && canManage ? (
									<button
										type="button"
										onClick={handleAddMembers}
										aria-label={t("group_settings.roles.detail.remove_member_aria", {
											name,
										})}
										className="inline-flex size-7 shrink-0 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-foreground"
									>
										<X className="size-4" strokeWidth={2} aria-hidden />
									</button>
								) : null}
							</div>
						);
					})}
				</div>
			)}
		</div>
	);
}
