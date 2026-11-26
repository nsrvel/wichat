"use client";

import { CalendarDays } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { HexColorField } from "@/components/ui/color-picker";
import { Input } from "@/components/ui/input";
import { UserAvatar } from "@/components/ui/user-avatar";
import { authFieldLabelClass, authInputClass } from "@/lib/form-field-styles";
import {
	settingsFormFooterBarClass,
	settingsSectionDividerLineClass,
	settingsSectionDividerWrapClass,
	settingsSectionGapClass,
} from "@/lib/settings-chrome";
import { UserAvatarUpload } from "@/components/user/user-avatar-upload";
import { UserStatusNotePill } from "@/components/user/user-status-note-pill";
import { translateError } from "@/integrations/i18n/translate-error";
import { UserProfileCardShell } from "@/components/user/user-profile-card-shell";
import { cn } from "@/lib/utils";
import {
	USER_PROFILE_AVATAR_OVERLAP_PX,
	USER_PROFILE_AVATAR_RING_CLASS,
	USER_PROFILE_AVATAR_SIZE,
	USER_PROFILE_BIO_CLASS,
	USER_PROFILE_CARD_WIDTH_PX,
	USER_PROFILE_DISPLAY_NAME_CLASS,
	USER_PROFILE_META_CLASS,
} from "@/lib/user-profile-card-layout";
import type { UserPresenceStatus } from "@/lib/user-status";
import { useSaveMyProfileMutation } from "@/services/user/user.mutation";

const DEFAULT_BANNER_COLOR = "#313338";
const descriptionInputClass = cn(authInputClass, "min-h-[6.5rem] resize-y py-2.5 sm:min-h-[7rem]");

type UserSettingsProfilesSectionProps = {
	displayName: string;
	username: string;
	avatarUrl: string;
	initials: string;
	status: UserPresenceStatus;
	statusNoteLabel?: string;
	createdAt?: string;
	initialPronouns?: string;
	initialBio?: string;
	initialBannerColor?: string;
};

export function UserSettingsProfilesSection({
	displayName,
	username,
	avatarUrl,
	initials,
	status,
	statusNoteLabel = "",
	createdAt = "",
	initialPronouns = "",
	initialBio = "",
	initialBannerColor = "",
}: UserSettingsProfilesSectionProps) {
	const { t, i18n } = useTranslation();
	const saveProfile = useSaveMyProfileMutation();
	const resolvedInitialBanner = initialBannerColor.trim() || DEFAULT_BANNER_COLOR;

	const [avatarFile, setAvatarFile] = useState<File | null>(null);
	const [avatarCleared, setAvatarCleared] = useState(false);
	const [nameDraft, setNameDraft] = useState(displayName);
	const [pronounsDraft, setPronounsDraft] = useState(initialPronouns);
	const [bannerColor, setBannerColor] = useState(resolvedInitialBanner);
	const [bioDraft, setBioDraft] = useState(initialBio);

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
		pronounsDraft !== initialPronouns ||
		bannerColor !== resolvedInitialBanner ||
		bioDraft !== initialBio;

	const handleReset = () => {
		setAvatarFile(null);
		setAvatarCleared(false);
		setNameDraft(displayName);
		setPronounsDraft(initialPronouns);
		setBannerColor(resolvedInitialBanner);
		setBioDraft(initialBio);
	};

	const handleSave = async () => {
		try {
			const saved = await saveProfile.mutateAsync({
				display_name: nameDraft.trim() || displayName,
				pronouns: pronounsDraft.trim(),
				bio: bioDraft.trim(),
				banner_color: bannerColor === DEFAULT_BANNER_COLOR ? "" : bannerColor,
				clear_avatar: avatarCleared && !avatarFile,
				avatarFile,
			});
			setAvatarFile(null);
			setAvatarCleared(false);
			setNameDraft(saved.display_name || displayName);
			setPronounsDraft(saved.pronouns ?? "");
			setBioDraft(saved.bio ?? "");
			setBannerColor(saved.banner_color?.trim() || DEFAULT_BANNER_COLOR);
			toast.success(t("user_settings.profiles.save_success"));
		} catch (err) {
			toast.error(translateError(t, err));
		}
	};

	const effectiveAvatarSrc = avatarPreviewUrl ?? (avatarCleared ? "" : avatarUrl);
	const trimmedPronouns = pronounsDraft.trim();
	const formattedMemberSince = useMemo(() => {
		const raw = createdAt.trim();
		const parsed = raw ? new Date(raw) : new Date();
		const date = Number.isNaN(parsed.getTime()) ? new Date() : parsed;
		return new Intl.DateTimeFormat(i18n.language, {
			month: "short",
			day: "numeric",
			year: "numeric",
		}).format(date);
	}, [createdAt, i18n.language]);

	return (
		<div className="flex flex-col">
			<div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_320px] lg:items-start">
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

					<div className={cn(settingsSectionGapClass, "space-y-4")}>
						<div className="space-y-2">
							<label
								htmlFor="user-settings-display-name"
								className={cn(authFieldLabelClass, "block")}
							>
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

					<div className={settingsSectionDividerWrapClass} role="presentation">
						<div className={settingsSectionDividerLineClass} />
					</div>

					<div className="space-y-2">
						<label
							htmlFor="user-settings-banner-color"
							className={cn(authFieldLabelClass, "block")}
						>
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

					<div className={settingsSectionDividerWrapClass} role="presentation">
						<div className={settingsSectionDividerLineClass} />
					</div>

					<div className="space-y-1.5">
						<p className={authFieldLabelClass}>{t("user_settings.profiles.member_since")}</p>
						<div className="flex items-center gap-2 text-sm text-muted-foreground">
							<CalendarDays className="size-4 shrink-0 opacity-80" aria-hidden />
							<span className="font-medium text-foreground/90">{formattedMemberSince}</span>
						</div>
					</div>
				</div>

				{/* Right column — sticky live preview card matching UserProfilePopover */}
				<div className="lg:sticky lg:top-0">
					<div className="mx-auto w-full max-w-[320px] overflow-hidden rounded-2xl border border-border bg-card shadow-sm lg:mx-0">
						<UserProfileCardShell
							widthPx={USER_PROFILE_CARD_WIDTH_PX}
							bannerColor={bannerColor}
							className="w-full"
						>
							<div
								className="flex items-center gap-2.5"
								style={{ marginTop: -USER_PROFILE_AVATAR_OVERLAP_PX }}
							>
								<UserAvatar
									src={effectiveAvatarSrc}
									initials={initials}
									size={USER_PROFILE_AVATAR_SIZE}
									status={status}
									ringClassName={USER_PROFILE_AVATAR_RING_CLASS}
								/>
								<UserStatusNotePill label={statusNoteLabel} />
							</div>
							<div className="mt-3.5 min-w-0">
								<p className={USER_PROFILE_DISPLAY_NAME_CLASS}>
									{nameDraft.trim() || displayName}
								</p>
								{username || trimmedPronouns ? (
									<p className={USER_PROFILE_META_CLASS}>
										{username ? `@${username}` : ""}
										{username && trimmedPronouns ? " • " : ""}
										{trimmedPronouns}
									</p>
								) : null}
								{bioDraft.trim() ? (
									<p className={USER_PROFILE_BIO_CLASS}>{bioDraft.trim()}</p>
								) : null}
							</div>
						</UserProfileCardShell>
					</div>
				</div>
			</div>

			{/* Floating unsaved changes bar */}
			{dirty ? (
				<div className="sticky bottom-3 z-20 mt-6">
					<div className={cn(settingsFormFooterBarClass, "shadow-lg")}>
						<p className="min-w-0 flex-1 truncate text-xs font-medium text-muted-foreground sm:text-sm">
							{t("user_settings.profiles.unsaved_hint")}
						</p>
						<div className="flex shrink-0 items-center gap-2">
							<Button
								type="button"
								variant="ghost"
								size="sm"
								disabled={saveProfile.isPending}
								onClick={handleReset}
							>
								{t("user_settings.profiles.reset")}
							</Button>
							<Button
								type="button"
								size="sm"
								disabled={saveProfile.isPending}
								onClick={() => void handleSave()}
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
