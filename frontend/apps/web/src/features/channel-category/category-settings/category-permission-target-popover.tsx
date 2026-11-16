"use client";

import type { ChannelCategoryPermissionOverride } from "@wichat/contracts/chat/channel-category";
import type { GroupMember } from "@wichat/contracts/group/member";
import type { GroupRole } from "@wichat/contracts/group/role";
import { Plus, Search, X } from "lucide-react";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { UserAvatar } from "@/components/ui/user-avatar";
import { hasCustomRoleColor } from "@/lib/member-role-display";
import { userInitials } from "@/lib/user-initials";

type CategoryPermissionTargetPopoverProps = {
	roles: readonly GroupRole[];
	members: readonly GroupMember[];
	onSelectTarget: (item: ChannelCategoryPermissionOverride) => void;
};

export function CategoryPermissionTargetPopover({
	roles,
	members,
	onSelectTarget,
}: CategoryPermissionTargetPopoverProps) {
	const { t } = useTranslation();
	const [open, setOpen] = useState(false);
	const [search, setSearch] = useState("");

	const handleOpenChange = (nextOpen: boolean) => {
		setOpen(nextOpen);
		if (!nextOpen) {
			setSearch("");
		}
	};

	const q = search.trim().toLowerCase();

	const filteredRoles = useMemo(() => {
		if (!q) {
			return roles;
		}
		return roles.filter((r) => r.name.toLowerCase().includes(q));
	}, [q, roles]);

	const filteredMembers = useMemo(() => {
		if (!q) {
			return members;
		}
		return members.filter(
			(m) =>
				(m.display_name || "").toLowerCase().includes(q) || m.username.toLowerCase().includes(q),
		);
	}, [members, q]);

	const hasAnyAvailable = roles.length > 0 || members.length > 0;
	const hasMatches = filteredRoles.length > 0 || filteredMembers.length > 0;

	return (
		<Popover open={open} onOpenChange={handleOpenChange}>
			<PopoverTrigger
				render={
					<Button
						type="button"
						variant="ghost"
						size="sm"
						className="h-8 w-full justify-between rounded-lg px-2 text-xs font-semibold text-foreground hover:bg-accent/70"
						aria-label={t("channel_category.settings.add_target_label")}
					/>
				}
			>
				<span>{t("channel_category.settings.roles_members_label")}</span>
				<Plus className="size-3.5 text-muted-foreground" aria-hidden />
			</PopoverTrigger>
			<PopoverContent
				align="start"
				sideOffset={6}
				className="w-72 overflow-hidden border-border bg-card p-0 text-card-foreground shadow-elevated"
			>
				<div className="border-b border-border/70 p-2">
					<div className="relative flex items-center">
						<Input
							type="text"
							value={search}
							onChange={(e) => setSearch(e.target.value)}
							placeholder={t("channel_category.settings.search_targets_placeholder")}
							aria-label={t("channel_category.settings.search_targets_placeholder")}
							autoComplete="off"
							className="h-9 w-full rounded-md border-0 bg-[#f5f5f6] pl-3.5 pr-9 text-[13px] text-foreground shadow-none placeholder:text-muted-foreground focus-visible:border-0 focus-visible:bg-[#f5f5f6] focus-visible:ring-0 dark:bg-[#242424] dark:focus-visible:bg-[#242424]"
						/>
						{search ? (
							<button
								type="button"
								onClick={() => setSearch("")}
								aria-label={t("expression.clear_search")}
								className="absolute right-2.5 flex size-5 items-center justify-center rounded text-muted-foreground hover:text-foreground"
							>
								<X className="size-3.5" aria-hidden />
							</button>
						) : (
							<Search
								className="pointer-events-none absolute right-3 size-4 text-muted-foreground"
								aria-hidden
							/>
						)}
					</div>
				</div>

				<div className="thin-scroll max-h-64 space-y-2 overflow-y-auto p-1.5">
					{!hasAnyAvailable ? (
						<p className="px-2 py-5 text-center text-xs text-muted-foreground">
							{t("channel_category.settings.add_target_empty")}
						</p>
					) : !hasMatches ? (
						<p className="px-2 py-5 text-center text-xs text-muted-foreground">
							{t("channel_category.settings.no_targets_match")}
						</p>
					) : (
						<>
							{filteredRoles.length > 0 ? (
								<div className="space-y-0.5">
									<p className="px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
										{t("channel_category.settings.roles_heading")}
									</p>
									{filteredRoles.map((role) => {
										const hasColor = hasCustomRoleColor(role.color);
										return (
											<button
												key={role.id}
												type="button"
												onClick={() => {
													onSelectTarget({
														target_type: "role",
														target_id: role.id,
														target_name: role.name,
														target_color: role.color,
														target_avatar_url: "",
														is_everyone: role.is_default || role.slug === "everyone",
														allow_permissions: [],
														deny_permissions: [],
													});
													setOpen(false);
												}}
												className="flex h-8 w-full items-center gap-2 rounded-lg px-2 text-left text-xs transition-colors hover:bg-accent"
											>
												<span className="flex size-5 shrink-0 items-center justify-center">
													<span
														className="size-2.5 rounded-full bg-muted-foreground/60"
														style={hasColor ? { backgroundColor: role.color } : undefined}
													/>
												</span>
												<span className="min-w-0 flex-1 truncate font-medium text-foreground">
													{role.name}
												</span>
											</button>
										);
									})}
								</div>
							) : null}

							{filteredMembers.length > 0 ? (
								<div className="space-y-0.5">
									<p className="px-2 py-1 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
										{t("channel_category.settings.members_heading")}
									</p>
									{filteredMembers.map((member) => {
										const displayName = member.display_name || member.username;
										return (
											<button
												key={member.user_id}
												type="button"
												onClick={() => {
													onSelectTarget({
														target_type: "member",
														target_id: member.user_id,
														target_name: displayName,
														target_color: "",
														target_avatar_url: member.avatar_url ?? "",
														is_everyone: false,
														allow_permissions: [],
														deny_permissions: [],
													});
													setOpen(false);
												}}
												className="flex h-8 w-full items-center gap-2 rounded-lg px-2 text-left text-xs transition-colors hover:bg-accent"
											>
												<UserAvatar
													src={member.avatar_url}
													initials={userInitials(displayName)}
													size="xs"
													className="size-5 text-[10px]"
												/>
												<span className="min-w-0 flex-1 truncate font-medium text-foreground">
													{displayName}
												</span>
											</button>
										);
									})}
								</div>
							) : null}
						</>
					)}
				</div>
			</PopoverContent>
		</Popover>
	);
}
