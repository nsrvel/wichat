"use client";

import { Search } from "lucide-react";
import { useId } from "react";
import { useTranslation } from "react-i18next";
import { authInputClass } from "@/features/auth/components/auth-shell";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

type GroupSettingsRolesToolbarProps = {
	search: string;
	onSearchChange: (value: string) => void;
	onCreateRole: () => void;
	canManage: boolean;
};

export function GroupSettingsRolesToolbar({
	search,
	onSearchChange,
	onCreateRole,
	canManage,
}: GroupSettingsRolesToolbarProps) {
	const { t } = useTranslation();
	const searchId = useId();

	return (
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
					placeholder={t("group_settings.roles.search_placeholder")}
					className={cn(authInputClass, "pl-10 pr-3 sm:pl-10 sm:pr-3.5")}
					autoComplete="off"
				/>
			</label>
			{canManage ? (
				<Button
					type="button"
					className="h-10 w-fit shrink-0 self-end sm:self-auto"
					onClick={onCreateRole}
				>
					{t("group_settings.roles.create_button")}
				</Button>
			) : null}
		</div>
	);
}
