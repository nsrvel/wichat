"use client";

import { Menu } from "@base-ui/react/menu";
import { ArrowDownUp, ChevronDown, ChevronUp, Search } from "lucide-react";
import { useId, useState } from "react";
import { useTranslation } from "react-i18next";
import { authInputClass } from "@/lib/form-field-styles";
import { GroupSettingsMembersPruneDialog } from "@/features/group-settings/components/members/group-settings-members-prune-dialog";
import type {
	GroupSettingsMemberRow,
	MembersSortKey,
} from "@/features/group-settings/lib/group-settings-members-view";
import { MenuRadioControl } from "@/components/ui/menu-radio-control";
import {
	menuPopupClass,
	menuPositionerClass,
	menuRadioItemClass,
} from "@/lib/menu-chrome";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

const SORT_OPTIONS: MembersSortKey[] = [
	"name_asc",
	"name_desc",
	"member_since_desc",
	"member_since_asc",
	"last_updated_desc",
	"last_updated_asc",
	"role_asc",
];

const toolbarActionButtonClass = cn(
	authInputClass,
	"inline-flex h-10 min-h-10 w-fit max-w-none shrink-0 items-center gap-2 px-3 text-sm font-medium text-foreground",
	"hover:bg-field/80 sm:px-3.5",
);

type GroupSettingsMembersToolbarProps = {
	search: string;
	onSearchChange: (value: string) => void;
	sort: MembersSortKey;
	onSortChange: (value: MembersSortKey) => void;
	groupName: string;
	rows: GroupSettingsMemberRow[];
	canPrune: boolean;
};

export function GroupSettingsMembersToolbar({
	search,
	onSearchChange,
	sort,
	onSortChange,
	groupName,
	rows,
	canPrune,
}: GroupSettingsMembersToolbarProps) {
	const { t } = useTranslation();
	const searchId = useId();
	const [sortOpen, setSortOpen] = useState(false);
	const [pruneOpen, setPruneOpen] = useState(false);

	return (
		<>
			<div className="flex flex-col gap-2 sm:flex-row sm:items-center">
				<label className="relative block min-w-0 flex-1" htmlFor={searchId}>
					<Search
						className="pointer-events-none absolute top-1/2 left-3.5 z-[1] size-4 -translate-y-1/2 text-muted-foreground"
						aria-hidden
					/>
					<Input
						id={searchId}
						value={search}
						onChange={(e) => onSearchChange(e.target.value)}
						placeholder={t("group_settings.members.search_placeholder")}
						className={cn(authInputClass, "pl-10 pr-3 sm:pl-10 sm:pr-3.5")}
						autoComplete="off"
					/>
				</label>

				<div className="flex w-fit max-w-full shrink-0 items-center gap-2 self-end sm:self-auto">
					<Menu.Root open={sortOpen} onOpenChange={setSortOpen}>
						<Menu.Trigger
							nativeButton
							render={
								<button
									type="button"
									className={cn(toolbarActionButtonClass, sortOpen && "ring-1 ring-ring/40")}
									aria-haspopup="menu"
									aria-expanded={sortOpen}
									aria-label={t("group_settings.members.sort_label")}
								/>
							}
						>
							<ArrowDownUp className="size-4 shrink-0 text-muted-foreground" aria-hidden />
							<span>{t("group_settings.members.sort_button")}</span>
							{sortOpen ? (
								<ChevronUp
									className="ml-0.5 size-3.5 text-muted-foreground"
									strokeWidth={2.4}
									aria-hidden
								/>
							) : (
								<ChevronDown
									className="ml-0.5 size-3.5 text-muted-foreground"
									strokeWidth={2.4}
									aria-hidden
								/>
							)}
						</Menu.Trigger>
						<Menu.Portal>
							<Menu.Positioner
								className={menuPositionerClass}
								side="bottom"
								align="end"
								sideOffset={4}
							>
								<Menu.Popup className={cn(menuPopupClass, "min-w-[14rem]")}>
									<Menu.RadioGroup
										value={sort}
										onValueChange={(next) => {
											if (SORT_OPTIONS.includes(next as MembersSortKey)) {
												onSortChange(next as MembersSortKey);
											}
										}}
									>
										{SORT_OPTIONS.map((key) => (
											<Menu.RadioItem
												key={key}
												value={key}
												className={cn(menuRadioItemClass, "group/radio")}
												closeOnClick
											>
												<span className="min-w-0 flex-1 pr-1">
													{t(`group_settings.members.sort.${key}`)}
												</span>
												<MenuRadioControl />
											</Menu.RadioItem>
										))}
									</Menu.RadioGroup>
								</Menu.Popup>
							</Menu.Positioner>
						</Menu.Portal>
					</Menu.Root>

					{canPrune ? (
						<button
							type="button"
							className={cn(
								toolbarActionButtonClass,
								"text-destructive hover:text-destructive hover:bg-destructive/10",
							)}
							onClick={() => setPruneOpen(true)}
						>
							{t("group_settings.members.prune_button")}
						</button>
					) : null}
				</div>
			</div>

			<GroupSettingsMembersPruneDialog
				open={pruneOpen}
				onOpenChange={setPruneOpen}
				groupName={groupName}
				rows={rows}
			/>
		</>
	);
}
