"use client";

import { Search } from "lucide-react";
import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Input } from "@/components/ui/input";
import { authInputClass } from "@/lib/form-field-styles";
import { SettingsToggle } from "@/components/ui/settings-toggle";
import {
	defaultPermissionsForRole,
	ROLE_PERMISSION_GROUPS,
	type GroupSettingsRoleRow,
} from "@/features/group-settings/lib/group-settings-roles-view";
import { cn } from "@/lib/utils";

type RoleDetailPermissionsTabProps = {
	role?: GroupSettingsRoleRow;
	canManage: boolean;
};

export function RoleDetailPermissionsTab({ canManage }: RoleDetailPermissionsTabProps) {
	const { t } = useTranslation();
	const [search, setSearch] = useState("");
	const [permissions, setPermissions] = useState<Record<string, boolean>>(() =>
		defaultPermissionsForRole(),
	);

	const editable = canManage;

	const filteredGroups = useMemo(() => {
		const q = search.trim().toLowerCase();
		if (!q) {
			return ROLE_PERMISSION_GROUPS;
		}
		return ROLE_PERMISSION_GROUPS.map((group) => ({
			...group,
			items: group.items.filter((item) => {
				const title = t(item.titleKey).toLowerCase();
				const desc = t(item.descriptionKey).toLowerCase();
				return title.includes(q) || desc.includes(q);
			}),
		})).filter((group) => group.items.length > 0);
	}, [search, t]);

	const handleClearAll = () => {
		if (!editable) {
			return;
		}
		setPermissions((prev) => {
			const next: Record<string, boolean> = {};
			for (const key of Object.keys(prev)) {
				next[key] = false;
			}
			return next;
		});
	};

	return (
		<div className="flex flex-col gap-6 pt-5">
			<div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
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
						placeholder={t("group_settings.roles.detail.permissions_search_placeholder")}
						className={cn(authInputClass, "pl-9")}
					/>
				</div>

				<button
					type="button"
					disabled={!editable}
					onClick={handleClearAll}
					className="shrink-0 self-end text-xs font-medium text-primary transition-opacity hover:underline disabled:pointer-events-none disabled:opacity-40 sm:self-center"
				>
					{t("group_settings.roles.detail.permissions_clear")}
				</button>
			</div>

			{filteredGroups.length === 0 ? (
				<div className="rounded-xl border border-dashed border-border/80 bg-field/30 px-4 py-8 text-center text-sm text-muted-foreground">
					{t("group_settings.roles.detail.permissions_empty")}
				</div>
			) : (
				<div className="flex flex-col gap-7">
					{filteredGroups.map((group, groupIndex) => (
						<div key={group.id} className="flex flex-col gap-4">
							{groupIndex > 0 ? <div className="border-t border-border pb-1" /> : null}
							<h3 className="text-xs font-semibold tracking-wider text-muted-foreground uppercase">
								{t(group.titleKey)}
							</h3>
							<div className="flex flex-col divide-y divide-border/70">
								{group.items.map((item) => {
									const checked = Boolean(permissions[item.id]);
									return (
										<div
											key={item.id}
											className="flex items-start justify-between gap-4 py-3.5 first:pt-1 last:pb-1"
										>
											<div className="flex min-w-0 flex-col gap-1">
												<span className="text-sm font-medium text-foreground">
													{t(item.titleKey)}
												</span>
												<p className="text-xs leading-relaxed text-muted-foreground">
													{t(item.descriptionKey)}
												</p>
											</div>
											<SettingsToggle
												checked={checked}
												disabled={!editable}
												onCheckedChange={(next) =>
													setPermissions((prev) => ({
														...prev,
														[item.id]: next,
													}))
												}
											/>
										</div>
									);
								})}
							</div>
						</div>
					))}
				</div>
			)}
		</div>
	);
}
