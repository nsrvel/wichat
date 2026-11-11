"use client";

import { Menu } from "@base-ui/react/menu";
import { Check, ChevronDown, ChevronUp } from "lucide-react";
import { useId, useMemo, useState } from "react";
import { Trans, useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { WiChatDialog, WiChatDialogBody, WiChatDialogFooter } from "@/components/ui/wichat-dialog";
import { authInputClass } from "@/features/auth/components/auth-shell";
import {
	countPruneKickCandidates,
	pruneInactiveDayCount,
	type GroupSettingsMemberRow,
	type MembersPruneInactiveKey,
	type PruneDialogRoleKey,
} from "@/features/group-settings/lib/group-settings-members-view";
import { MenuRadioControl } from "@/features/shell/components/group-rail-menu/menu-radio-control";
import {
	menuCheckboxBoxClass,
	menuCheckboxItemClass,
	menuDropdownTriggerChevronClass,
	menuItemClass,
	menuPopupClass,
	menuPositionerClass,
	menuRadioItemClass,
} from "@/features/shell/components/group-rail-menu/menu-chrome";
import { cn } from "@/lib/utils";

const PRUNE_INACTIVE_OPTIONS: MembersPruneInactiveKey[] = ["inactive_7d", "inactive_30d"];
const PRUNE_ROLE_OPTIONS: PruneDialogRoleKey[] = ["member", "owner"];

const PRUNE_DIALOG_TITLE_ID = "group-settings-prune-members-title";

type GroupSettingsMembersPruneDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	groupName: string;
	rows: GroupSettingsMemberRow[];
};

export function GroupSettingsMembersPruneDialog({
	open,
	onOpenChange,
	groupName,
	rows,
}: GroupSettingsMembersPruneDialogProps) {
	const { t } = useTranslation();
	const lastSeenLegendId = useId();
	const rolesLabelId = useId();
	const [inactive, setInactive] = useState<MembersPruneInactiveKey>("inactive_7d");
	const [alsoRoles, setAlsoRoles] = useState<PruneDialogRoleKey[]>([]);
	const [rolesMenuOpen, setRolesMenuOpen] = useState(false);

	const kickCount = useMemo(
		() => countPruneKickCandidates(rows, inactive, alsoRoles),
		[rows, inactive, alsoRoles],
	);
	const inactiveDays = pruneInactiveDayCount(inactive);

	const rolesTriggerLabel =
		alsoRoles.length === 0
			? t("group_settings.members.prune_dialog.roles_placeholder")
			: alsoRoles.map((r) => t(`invite.role_${r}`)).join(", ");

	return (
		<WiChatDialog
			open={open}
			onOpenChange={onOpenChange}
			titleId={PRUNE_DIALOG_TITLE_ID}
			elevated
			className="max-w-md"
		>
			<WiChatDialogBody className="pb-2 pt-10 sm:pt-11">
				<h2
					id={PRUNE_DIALOG_TITLE_ID}
					className="text-pretty text-left text-xl font-bold tracking-tight text-foreground sm:text-2xl"
				>
					{t("group_settings.members.prune_dialog.title", { name: groupName })}
				</h2>

				<div className="mt-6 space-y-6">
					<fieldset className="space-y-3">
						<legend id={lastSeenLegendId} className="text-sm font-semibold text-foreground">
							{t("group_settings.members.prune_dialog.last_seen")}
						</legend>
						<div role="radiogroup" aria-labelledby={lastSeenLegendId} className="space-y-1">
							{PRUNE_INACTIVE_OPTIONS.map((key) => {
								const selected = inactive === key;
								return (
									<button
										key={key}
										type="button"
										role="radio"
										aria-checked={selected}
										className={cn(menuRadioItemClass, "group/radio w-full")}
										data-checked={selected ? "" : undefined}
										onClick={() => setInactive(key)}
									>
										<span className="min-w-0 flex-1 pr-1 text-left text-sm">
											{t(`group_settings.members.prune_dialog.last_seen_${key}`)}
										</span>
										<MenuRadioControl />
									</button>
								);
							})}
						</div>
					</fieldset>

					<div className="space-y-2">
						<p id={rolesLabelId} className="text-sm font-semibold text-foreground">
							{t("group_settings.members.prune_dialog.roles_label")}
						</p>
						<Menu.Root open={rolesMenuOpen} onOpenChange={setRolesMenuOpen}>
							<Menu.Trigger
								nativeButton
								render={
									<button
										type="button"
										aria-labelledby={rolesLabelId}
										className={cn(
											authInputClass,
											"flex h-10 w-full items-center justify-between gap-2 text-left text-sm",
										)}
									/>
								}
							>
								<span
									className={cn(
										"min-w-0 truncate",
										alsoRoles.length === 0 ? "text-muted-foreground" : "text-field-foreground",
									)}
								>
									{rolesTriggerLabel}
								</span>
								{rolesMenuOpen ? (
									<ChevronUp
										className={menuDropdownTriggerChevronClass}
										strokeWidth={2.4}
										aria-hidden
									/>
								) : (
									<ChevronDown
										className={menuDropdownTriggerChevronClass}
										strokeWidth={2.4}
										aria-hidden
									/>
								)}
							</Menu.Trigger>
							<Menu.Portal>
								<Menu.Positioner
									className={menuPositionerClass}
									side="bottom"
									align="start"
									sideOffset={4}
								>
									<Menu.Popup className={cn(menuPopupClass, "min-w-[var(--anchor-width)]")}>
										{PRUNE_ROLE_OPTIONS.map((role) => {
											const checked = alsoRoles.includes(role);
											return (
												<Menu.CheckboxItem
													key={role}
													className={cn(menuCheckboxItemClass, "group/checkbox")}
													checked={checked}
													closeOnClick={false}
													onCheckedChange={(next) => {
														setAlsoRoles((prev) =>
															next ? [...new Set([...prev, role])] : prev.filter((r) => r !== role),
														);
													}}
												>
													<span className="min-w-0 flex-1 pr-1">{t(`invite.role_${role}`)}</span>
													<span className={menuCheckboxBoxClass} aria-hidden>
														<Menu.CheckboxItemIndicator
															keepMounted
															className="flex items-center justify-center opacity-0 data-[checked]:opacity-100"
														>
															<Check className="h-3 w-3" strokeWidth={2.5} />
														</Menu.CheckboxItemIndicator>
													</span>
												</Menu.CheckboxItem>
											);
										})}
										<Menu.Separator className="mx-2.5 my-2 h-px bg-border/70" />
										<Menu.Item
											className={menuItemClass}
											closeOnClick
											onClick={() => setAlsoRoles([])}
										>
											{t("group_settings.members.prune_dialog.roles_clear")}
										</Menu.Item>
									</Menu.Popup>
								</Menu.Positioner>
							</Menu.Portal>
						</Menu.Root>
					</div>

					<p className="text-sm leading-relaxed text-muted-foreground">
						<Trans
							i18nKey={
								alsoRoles.includes("owner")
									? "group_settings.members.prune_dialog.summary_with_owner"
									: "group_settings.members.prune_dialog.summary"
							}
							values={{ count: kickCount, days: inactiveDays }}
							components={{
								strong: <span className="font-semibold text-foreground" />,
							}}
						/>
					</p>
				</div>
			</WiChatDialogBody>

			<WiChatDialogFooter>
				<div className="flex gap-3">
					<Button
						type="button"
						variant="secondary"
						className="h-10 min-h-10 flex-1 rounded-[10px]"
						onClick={() => onOpenChange(false)}
					>
						{t("group_settings.members.prune_dialog.cancel")}
					</Button>
					<Button
						type="button"
						className="h-10 min-h-10 flex-1 rounded-[10px]"
						disabled={kickCount === 0}
						onClick={() => {
							// ponytail: POST prune RPC when bulk-remove members ships.
							onOpenChange(false);
						}}
					>
						{t("group_settings.members.prune_dialog.confirm")}
					</Button>
				</div>
			</WiChatDialogFooter>
		</WiChatDialog>
	);
}
