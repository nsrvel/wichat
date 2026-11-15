"use client";

import { Search } from "lucide-react";
import { useId, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Input } from "@/components/ui/input";
import { SettingsToggle } from "@/components/ui/settings-toggle";
import { groupSettingsToolbarSearchInputClass } from "@/features/group-settings/lib/group-settings-chrome";
import { ROLE_PERMISSION_GROUPS } from "@/features/group-settings/lib/group-settings-roles-view";
import { authInputClass } from "@/lib/form-field-styles";
import { cn } from "@/lib/utils";

type RoleDetailPermissionsTabProps = {
	canManage: boolean;
	canGrantAdministrator: boolean;
	permissions: Record<string, boolean>;
	onPermissionsChange: (next: Record<string, boolean>) => void;
};

export function RoleDetailPermissionsTab({
	canManage,
	canGrantAdministrator,
	permissions,
	onPermissionsChange,
}: RoleDetailPermissionsTabProps) {
	const { t } = useTranslation();
	const searchId = useId();
	const [search, setSearch] = useState("");

	const editable = canManage;
	const hasAnyEnabled = useMemo(
		() => Object.values(permissions).some(Boolean),
		[permissions],
	);

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
		const next: Record<string, boolean> = {};
		for (const key of Object.keys(permissions)) {
			next[key] = false;
		}
		onPermissionsChange(next);
	};

	return (
		<div className="flex flex-col gap-6 pt-5">
			<label className="relative block min-w-0" htmlFor={searchId}>
				<Search
					className="pointer-events-none absolute top-1/2 left-3.5 z-[1] size-4 -translate-y-1/2 text-muted-foreground"
					aria-hidden
				/>
				<Input
					id={searchId}
					value={search}
					onChange={(e) => setSearch(e.target.value)}
					placeholder={t("group_settings.roles.detail.permissions_search_placeholder")}
					className={cn(authInputClass, groupSettingsToolbarSearchInputClass)}
					autoComplete="off"
				/>
			</label>

			{filteredGroups.length === 0 ? (
				<div className="rounded-xl border border-dashed border-border/80 bg-field/30 px-4 py-8 text-center text-sm text-muted-foreground">
					{t("group_settings.roles.detail.permissions_empty")}
				</div>
			) : (
				<div className="flex flex-col gap-7">
					{filteredGroups.map((group, groupIndex) => (
						<div key={group.id} className="flex flex-col gap-4">
							{groupIndex > 0 ? <div className="border-t border-border pb-1" /> : null}
							<div className="flex items-center justify-between gap-3">
								<h3 className="text-xs font-semibold tracking-wider text-muted-foreground uppercase">
									{t(group.titleKey)}
								</h3>
								{groupIndex === 0 ? (
									<button
										type="button"
										disabled={!editable || !hasAnyEnabled}
										onClick={handleClearAll}
										className="shrink-0 text-xs font-medium text-primary transition-opacity hover:underline disabled:pointer-events-none disabled:opacity-40"
									>
										{t("group_settings.roles.detail.permissions_clear")}
									</button>
								) : null}
							</div>
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
												disabled={!editable || (item.id === "administrator" && !canGrantAdministrator)}
												onCheckedChange={(next) =>
													onPermissionsChange({
														...permissions,
														[item.id]: next,
													})
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
