"use client";

import { Check } from "lucide-react";
import { useId } from "react";
import { useTranslation } from "react-i18next";
import { FieldLabelContent } from "@/components/form/field-label";
import { HexColorField } from "@/components/ui/color-picker";
import { Input } from "@/components/ui/input";
import { authFieldLabelClass, authInputClass } from "@/lib/form-field-styles";
import { SettingsToggle } from "@/components/ui/settings-toggle";
import {
	DEFAULT_ROLE_COLOR,
	ROLE_COLOR_PRESETS,
	type GroupSettingsRoleRow,
} from "@/features/group-settings/lib/group-settings-roles-view";
import { cn } from "@/lib/utils";

type RoleDetailDisplayTabProps = {
	role: GroupSettingsRoleRow;
	canManage: boolean;
	roleName: string;
	onRoleNameChange: (value: string) => void;
	selectedColor: string;
	onSelectedColorChange: (value: string) => void;
	hoist: boolean;
	onHoistChange: (value: boolean) => void;
	mentionable: boolean;
	onMentionableChange: (value: boolean) => void;
};

export function RoleDetailDisplayTab({
	role,
	canManage,
	roleName,
	onRoleNameChange,
	selectedColor,
	onSelectedColorChange,
	hoist,
	onHoistChange,
	mentionable,
	onMentionableChange,
}: RoleDetailDisplayTabProps) {
	const { t } = useTranslation();
	const nameInputId = useId();
	const colorInputId = useId();

	const isEveryone = role.isDefault || role.slug === "everyone";
	const editable = canManage && !isEveryone;

	return (
		<div className="flex flex-col gap-6 pt-5">
			{isEveryone ? (
				<div className="rounded-xl border border-border/80 bg-field/40 px-4 py-3 text-xs leading-relaxed text-muted-foreground">
					{t("group_settings.roles.detail.everyone_display_readonly")}
				</div>
			) : null}

			{/* Role name */}
			<div className="flex flex-col gap-2">
				<label htmlFor={nameInputId} className={authFieldLabelClass}>
					<FieldLabelContent required={!isEveryone}>
						{t("group_settings.roles.detail.role_name_label")}
					</FieldLabelContent>
				</label>
				<Input
					id={nameInputId}
					value={roleName}
					disabled={!editable}
					placeholder={t("group_settings.roles.detail.role_name_placeholder")}
					onChange={(e) => onRoleNameChange(e.target.value)}
					className={authInputClass}
				/>
			</div>

			<div className="border-t border-border" />

			{/* Role color */}
			<div className="flex flex-col gap-3">
				<div className="flex flex-col gap-1">
					<label htmlFor={colorInputId} className={authFieldLabelClass}>
						<FieldLabelContent required={!isEveryone}>
							{t("group_settings.roles.detail.role_color_label")}
						</FieldLabelContent>
					</label>
					<p className="text-xs leading-relaxed text-muted-foreground">
						{t("group_settings.roles.detail.role_color_hint")}
					</p>
				</div>

				<div className="flex w-full max-w-[380px] flex-col gap-3 pt-1">
					{/* 2x10 Preset Palette */}
					<div className="grid grid-cols-5 gap-2 sm:grid-cols-10">
						{ROLE_COLOR_PRESETS.map((preset) => {
							const active = selectedColor.toLowerCase() === preset.toLowerCase();
							return (
								<button
									key={preset}
									type="button"
									disabled={!editable}
									aria-label={t("group_settings.roles.detail.preset_color_aria", {
										color: preset,
									})}
									aria-pressed={active}
									onClick={() => onSelectedColorChange(preset)}
									style={{ backgroundColor: preset }}
									className={cn(
										"flex h-7 w-full items-center justify-center rounded-md border border-black/10 transition-transform hover:scale-105 active:scale-95 disabled:pointer-events-none disabled:opacity-50 dark:border-white/10",
										active && "ring-2 ring-ring ring-offset-2 ring-offset-background",
									)}
								>
									{active ? (
										<Check
											className="size-3.5 text-white drop-shadow-xs"
											strokeWidth={2.8}
											aria-hidden
										/>
									) : null}
								</button>
							);
						})}
					</div>

					<HexColorField
						id={colorInputId}
						value={selectedColor}
						onChange={onSelectedColorChange}
						defaultColor={DEFAULT_ROLE_COLOR}
						disabled={!editable}
						resetLabel={t("group_settings.roles.detail.default_color_aria")}
						pickerAriaLabel={t("group_settings.roles.detail.role_color_label")}
						inputClassName={authInputClass}
					/>
				</div>
			</div>

			{!isEveryone ? (
				<>
					<div className="border-t border-border" />

					{/* Role display toggles */}
					<div className="flex flex-col gap-5">
						<div className="flex items-center justify-between gap-4">
							<span className="text-sm font-medium text-foreground">
								{t("group_settings.roles.detail.hoist_label")}
							</span>
							<SettingsToggle
								checked={hoist}
								disabled={!editable}
								onCheckedChange={onHoistChange}
							/>
						</div>

						<div className="flex flex-col gap-1.5">
							<div className="flex items-center justify-between gap-4">
								<span className="text-sm font-medium text-foreground">
									{t("group_settings.roles.detail.mentionable_label")}
								</span>
								<SettingsToggle
									checked={mentionable}
									disabled={!editable}
									onCheckedChange={onMentionableChange}
								/>
							</div>
							<p className="text-xs leading-relaxed text-muted-foreground">
								{t("group_settings.roles.detail.mentionable_hint")}
							</p>
						</div>
					</div>
				</>
			) : null}
		</div>
	);
}
