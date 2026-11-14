"use client";

import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { HexColorField } from "@/components/ui/color-picker";
import { Input } from "@/components/ui/input";
import { UserAvatar } from "@/components/ui/user-avatar";
import { authFieldLabelClass, authInputClass } from "@/features/auth/components/auth-shell";
import {
	groupSettingsFormFooterBarClass,
	groupSettingsProfileSectionDividerLineClass,
	groupSettingsProfileSectionDividerWrapClass,
	groupSettingsProfileSectionGapClass,
} from "@/features/group-settings/lib/group-settings-chrome";
import { UserAvatarUpload } from "@/features/user-settings/components/user-avatar-upload";
import { cn } from "@/lib/utils";
import type { UserPresenceStatus } from "@/lib/user-status";

const DEFAULT_BANNER_COLOR = "#313338";
const descriptionInputClass = cn(authInputClass, "min-h-[6.5rem] resize-y py-2.5 sm:min-h-[7rem]");

type UserSettingsProfilesSectionProps = {
	displayName: string;
	username: string;
	avatarUrl: string;
	initials: string;
	status: UserPresenceStatus;
};

export function UserSettingsProfilesSection({
	displayName,
	username,
	avatarUrl,
	initials,
	status,
}: UserSettingsProfilesSectionProps) {
	const { t } = useTranslation();
	const [avatarFile, setAvatarFile] = useState<File | null>(null);
	const [avatarCleared, setAvatarCleared] = useState(false);
	const [nameDraft, setNameDraft] = useState(displayName);
	const [pronounsDraft, setPronounsDraft] = useState("");
	const [bannerColor, setBannerColor] = useState(DEFAULT_BANNER_COLOR);
	const [bioDraft, setBioDraft] = useState("");

	const avatarPreviewUrl = useMemo(
		() => (avatarFile ? URL.createObjectURL(avatarFile) : null),
		[avatarFile],
	);

	useEffect(() => {
		if (!avatarPreviewUrl) {
			return;
		}
		return () => URL.revokeObjectURL(avatarPreviewUrl);
	}, [avatarPreviewUrl]);

	const dirty =
		avatarFile !== null ||
		avatarCleared ||
		nameDraft !== displayName ||
		pronounsDraft !== "" ||
		bannerColor !== DEFAULT_BANNER_COLOR ||
		bioDraft !== "";

	const handleReset = () => {
		setAvatarFile(null);
		setAvatarCleared(false);
		setNameDraft(displayName);
		setPronounsDraft("");
		setBannerColor(DEFAULT_BANNER_COLOR);
		setBioDraft("");
	};

	const effectiveAvatarSrc = avatarPreviewUrl ?? (avatarCleared ? "" : avatarUrl);
	const trimmedPronouns = pronounsDraft.trim();

	return (
		<div className="flex flex-col">
			<div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_272px] lg:items-start">
				{/* Left form column — styled identically to Group Settings -> Profile */}
				<div className="flex flex-col">
					<div id="user-settings-avatar">
						<UserAvatarUpload
						value={avatarFile}
						onChange={(file) => {
							setAvatarFile(file);
							if (file) {
								setAvatarCleared(false);
							}
						}}
						avatarUrl={avatarCleared ? "" : avatarUrl}
						initials={initials}
						onClearExisting={() => setAvatarCleared(true)}
					/>
					</div>

					<div className={cn(groupSettingsProfileSectionGapClass, "space-y-4")}>
						<div className="space-y-2">
							<label htmlFor="user-settings-display-name" className={cn(authFieldLabelClass, "block")}>
								{t("user_settings.profiles.display_name")}
							</label>
							<Input
								id="user-settings-display-name"
								value={nameDraft}
								onChange={(e) => setNameDraft(e.target.value)}
								maxLength={50}
								placeholder={t("user_settings.profiles.display_name_placeholder")}
								className={authInputClass}
							/>
						</div>

						<div className="space-y-2">
							<label htmlFor="user-settings-pronouns" className={cn(authFieldLabelClass, "block")}>
								{t("user_settings.profiles.pronouns")}
							</label>
							<Input
								id="user-settings-pronouns"
								value={pronounsDraft}
								onChange={(e) => setPronounsDraft(e.target.value)}
								maxLength={40}
								placeholder={t("user_settings.profiles.pronouns_placeholder")}
								className={authInputClass}
							/>
						</div>

						<div className="space-y-2">
							<label htmlFor="user-settings-bio" className={cn(authFieldLabelClass, "block")}>
								{t("user_settings.profiles.bio")}
							</label>
							<textarea
								id="user-settings-bio"
								value={bioDraft}
								onChange={(e) => setBioDraft(e.target.value)}
								rows={4}
								maxLength={190}
								placeholder={t("user_settings.profiles.bio_placeholder")}
								className={descriptionInputClass}
							/>
						</div>
					</div>

					<div className={groupSettingsProfileSectionDividerWrapClass} role="presentation">
						<div className={groupSettingsProfileSectionDividerLineClass} />
					</div>

					<div className="space-y-2">
						<label htmlFor="user-settings-banner-color" className={cn(authFieldLabelClass, "block")}>
							{t("user_settings.profiles.banner_color")}
						</label>
						<HexColorField
							id="user-settings-banner-color"
							value={bannerColor}
							onChange={setBannerColor}
							defaultColor={DEFAULT_BANNER_COLOR}
							resetLabel={t("user_settings.profiles.banner_color_reset")}
							pickerAriaLabel={t("user_settings.profiles.banner_color")}
							inputClassName={authInputClass}
						/>
					</div>
				</div>

				{/* Right column — sticky live preview card matching UserProfilePopover */}
				<div className="space-y-2 lg:sticky lg:top-0">
					<p className={authFieldLabelClass}>{t("user_settings.profiles.preview_label")}</p>
					<div className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
						<div
							className="h-[76px] transition-colors duration-150"
							style={{ backgroundColor: bannerColor }}
						/>
						<div className="px-5 pb-5">
							<div className="-mt-[40px]">
								<UserAvatar
									src={effectiveAvatarSrc}
									initials={initials}
									size="xl"
									status={status}
									ringClassName="ring-[5px] ring-card"
								/>
							</div>
							<div className="mt-3 min-w-0">
								<p className="truncate text-base font-bold leading-tight text-foreground">
									{nameDraft.trim() || displayName}
								</p>
								{username || trimmedPronouns ? (
									<p className="mt-0.5 truncate text-xs font-medium text-muted-foreground">
										{username ? `@${username}` : ""}
										{username && trimmedPronouns ? " • " : ""}
										{trimmedPronouns}
									</p>
								) : null}
								{bioDraft.trim() ? (
									<p className="mt-2.5 whitespace-pre-wrap break-words text-xs leading-relaxed text-muted-foreground">
										{bioDraft.trim()}
									</p>
								) : null}
							</div>
						</div>
					</div>
				</div>
			</div>

			{/* Floating unsaved changes bar */}
			{dirty ? (
				<div className="sticky bottom-3 z-20 mt-6">
					<div className={cn(groupSettingsFormFooterBarClass, "shadow-lg")}>
						<p className="min-w-0 flex-1 truncate text-xs font-medium text-muted-foreground sm:text-sm">
							{t("user_settings.profiles.unsaved_hint")}
						</p>
						<div className="flex shrink-0 items-center gap-2">
							<Button type="button" variant="ghost" size="sm" onClick={handleReset}>
								{t("user_settings.profiles.reset")}
							</Button>
							<Button
								type="button"
								size="sm"
								onClick={() => toast.info(t("shell.group_rail_menu.coming_soon"))}
							>
								{t("user_settings.profiles.save_changes")}
							</Button>
						</div>
					</div>
				</div>
			) : null}
		</div>
	);
}
