"use client";

import { Check, ImagePlus, Pencil } from "lucide-react";
import Image from "next/image";
import { useEffect, useId, useMemo, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { FieldLabelContent } from "@/components/form/field-label";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { authFieldLabelClass, authInputClass } from "@/lib/form-field-styles";
import { SettingsToggle } from "@/components/ui/settings-toggle";
import {
	DEFAULT_ROLE_COLOR,
	ROLE_COLOR_PRESETS,
	type GroupSettingsRoleRow,
} from "@/features/group-settings/lib/group-settings-roles-view";
import { cn } from "@/lib/utils";

const MAX_ROLE_ICON_BYTES = 256 * 1024;
const ROLE_ICON_ACCEPT = "image/png,image/jpeg,image/webp";

type RoleDetailDisplayTabProps = {
	role: GroupSettingsRoleRow;
	roleLabel: string;
	canManage: boolean;
};

export function RoleDetailDisplayTab({ role, roleLabel, canManage }: RoleDetailDisplayTabProps) {
	const { t } = useTranslation();
	const nameInputId = useId();
	const colorInputRef = useRef<HTMLInputElement>(null);
	const iconInputRef = useRef<HTMLInputElement>(null);

	const isEveryone = role.id === "everyone";
	const editable = canManage && !isEveryone;

	const [roleName, setRoleName] = useState(roleLabel);
	const [selectedColor, setSelectedColor] = useState(role.color || DEFAULT_ROLE_COLOR);
	const [customColor, setCustomColor] = useState(
		ROLE_COLOR_PRESETS.includes(role.color as (typeof ROLE_COLOR_PRESETS)[number])
			? "#5865f2"
			: role.color || "#5865f2",
	);
	const [iconFile, setIconFile] = useState<File | null>(null);
	const [hoist, setHoist] = useState(false);
	const [mentionable, setMentionable] = useState(false);

	const iconBlobUrl = useMemo(() => (iconFile ? URL.createObjectURL(iconFile) : null), [iconFile]);

	useEffect(() => {
		if (!iconBlobUrl) {
			return;
		}
		return () => URL.revokeObjectURL(iconBlobUrl);
	}, [iconBlobUrl]);

	const isDefaultSelected = selectedColor.toLowerCase() === DEFAULT_ROLE_COLOR.toLowerCase();
	const isPresetSelected = ROLE_COLOR_PRESETS.some(
		(preset) => preset.toLowerCase() === selectedColor.toLowerCase(),
	);
	const isCustomSelected = !isDefaultSelected && !isPresetSelected;

	const handleIconPick = (file: File | undefined) => {
		if (!file) {
			return;
		}
		if (file.size > MAX_ROLE_ICON_BYTES) {
			toast.error(t("group_settings.roles.detail.icon_too_large"));
			return;
		}
		setIconFile(file);
	};

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
					onChange={(e) => setRoleName(e.target.value)}
					className={authInputClass}
				/>
			</div>

			<div className="border-t border-border" />

			{/* Role color */}
			<div className="flex flex-col gap-3">
				<div className="flex flex-col gap-1">
					<span className={authFieldLabelClass}>
						<FieldLabelContent required={!isEveryone}>
							{t("group_settings.roles.detail.role_color_label")}
						</FieldLabelContent>
					</span>
					<p className="text-xs leading-relaxed text-muted-foreground">
						{t("group_settings.roles.detail.role_color_hint")}
					</p>
				</div>

				<div className="flex flex-wrap items-center gap-3 pt-1">
					{/* Default swatch */}
					<button
						type="button"
						disabled={!editable}
						aria-label={t("group_settings.roles.detail.default_color_aria")}
						aria-pressed={isDefaultSelected}
						onClick={() => setSelectedColor(DEFAULT_ROLE_COLOR)}
						style={{ backgroundColor: DEFAULT_ROLE_COLOR }}
						className={cn(
							"relative flex h-12 w-16 shrink-0 items-center justify-center rounded-lg border border-black/10 transition-transform active:scale-95 disabled:pointer-events-none disabled:opacity-50 dark:border-white/15",
							isDefaultSelected && "ring-2 ring-ring ring-offset-2 ring-offset-background",
						)}
					>
						{isDefaultSelected ? (
							<Check className="size-5 text-white drop-shadow-xs" strokeWidth={2.5} aria-hidden />
						) : null}
					</button>

					{/* Custom color picker swatch */}
					<div className="relative">
						<input
							ref={colorInputRef}
							type="color"
							value={customColor}
							disabled={!editable}
							onChange={(e) => {
								const next = e.target.value;
								setCustomColor(next);
								setSelectedColor(next);
							}}
							className="sr-only"
							tabIndex={-1}
							aria-hidden
						/>
						<button
							type="button"
							disabled={!editable}
							aria-label={t("group_settings.roles.detail.custom_color_aria")}
							aria-pressed={isCustomSelected}
							onClick={() => colorInputRef.current?.click()}
							style={isCustomSelected ? { backgroundColor: selectedColor } : undefined}
							className={cn(
								"relative flex h-12 w-16 shrink-0 items-center justify-center rounded-lg border border-border bg-field/50 transition-transform hover:bg-field active:scale-95 disabled:pointer-events-none disabled:opacity-50",
								isCustomSelected && "ring-2 ring-ring ring-offset-2 ring-offset-background",
							)}
						>
							{isCustomSelected ? (
								<Check className="size-5 text-white drop-shadow-xs" strokeWidth={2.5} aria-hidden />
							) : null}
							<Pencil
								className={cn(
									"absolute top-1.5 right-1.5 size-3.5",
									isCustomSelected ? "text-white drop-shadow-xs" : "text-foreground/80",
								)}
								strokeWidth={2.2}
								aria-hidden
							/>
						</button>
					</div>

					{/* 2x10 Preset Grid */}
					<div className="grid grid-cols-10 gap-2">
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
									onClick={() => setSelectedColor(preset)}
									style={{ backgroundColor: preset }}
									className={cn(
										"flex size-6 items-center justify-center rounded-full border border-black/10 transition-transform hover:scale-110 active:scale-95 disabled:pointer-events-none disabled:opacity-50 dark:border-white/10",
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
				</div>
			</div>

			<div className="border-t border-border" />

			{/* Role icon */}
			<div className="flex flex-col gap-3">
				<div className="flex flex-col gap-1">
					<span className={authFieldLabelClass}>
						{t("group_settings.roles.detail.role_icon_label")}
					</span>
					<p className="text-xs leading-relaxed text-muted-foreground">
						{t("group_settings.roles.detail.role_icon_hint")}
					</p>
				</div>

				<div className="flex items-center gap-4 pt-1">
					<input
						ref={iconInputRef}
						type="file"
						accept={ROLE_ICON_ACCEPT}
						disabled={!editable}
						className="sr-only"
						onChange={(e) => {
							handleIconPick(e.target.files?.[0]);
							e.target.value = "";
						}}
					/>
					<button
						type="button"
						disabled={!editable}
						onClick={() => iconInputRef.current?.click()}
						aria-label={t("group_settings.roles.detail.choose_image")}
						className="relative flex size-16 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-border bg-field/50 text-muted-foreground transition-colors hover:bg-field hover:text-foreground disabled:pointer-events-none disabled:opacity-50"
					>
						{iconBlobUrl ? (
							<Image
								src={iconBlobUrl}
								alt=""
								fill
								unoptimized
								sizes="64px"
								className="object-cover"
							/>
						) : (
							<ImagePlus className="size-6" strokeWidth={1.8} aria-hidden />
						)}
					</button>

					<div className="flex flex-wrap items-center gap-2">
						<Button
							type="button"
							variant="default"
							disabled={!editable}
							onClick={() => iconInputRef.current?.click()}
						>
							{t("group_settings.roles.detail.choose_image")}
						</Button>
						{iconFile ? (
							<Button
								type="button"
								variant="ghost"
								disabled={!editable}
								onClick={() => setIconFile(null)}
							>
								{t("group_settings.roles.detail.remove_image")}
							</Button>
						) : null}
					</div>
				</div>

				{/* Message appearance preview card */}
				<div className="mt-2 flex items-center gap-3 rounded-xl border border-border/80 bg-field/35 px-4 py-3.5">
					<div className="flex size-10 shrink-0 items-center justify-center rounded-full bg-primary/15 text-xs font-bold text-foreground ring-1 ring-border/60">
						W
					</div>
					<div className="min-w-0 flex-1">
						<div className="flex items-center gap-1.5">
							<span className="truncate text-sm font-semibold" style={{ color: selectedColor }}>
								{t("group_settings.roles.detail.preview_author")}
							</span>
							{iconBlobUrl ? (
								<span className="relative inline-flex size-4 shrink-0 overflow-hidden rounded-xs">
									<Image
										src={iconBlobUrl}
										alt=""
										fill
										unoptimized
										sizes="16px"
										className="object-cover"
									/>
								</span>
							) : (
								<ImagePlus
									className="size-4 shrink-0 text-muted-foreground"
									strokeWidth={1.8}
									aria-hidden
								/>
							)}
							<span className="text-[11px] text-muted-foreground">
								{t("group_settings.roles.detail.preview_time")}
							</span>
						</div>
						<p className="mt-0.5 text-sm text-foreground/90">
							{t("group_settings.roles.detail.preview_message")}
						</p>
					</div>
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
								onCheckedChange={setHoist}
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
									onCheckedChange={setMentionable}
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
