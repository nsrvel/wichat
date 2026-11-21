"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { ApiError } from "@wichat/api-client/api-error";
import type { Group } from "@wichat/contracts/group";
import { useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { FieldLabelContent } from "@/components/form/field-label";
import { TextField } from "@/components/form/text-field";
import { GroupIconUpload } from "@/components/group/group-icon-upload";
import { GroupProfileCardPreview } from "@/components/group/group-profile-card-preview";
import { HexColorField } from "@/components/ui/color-picker";
import {
	Form,
	FormControl,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { SettingsToggle } from "@/components/ui/settings-toggle";
import { GroupSettingsFormFooter } from "@/features/group-settings/components/group-settings-form-footer";
import { useGroupSettings } from "@/features/group-settings/group-settings-provider";
import {
	groupSettingsFormFooterFloatAlignClass,
	groupSettingsFormFooterFloatColumnClass,
	groupSettingsFormFooterFloatShellClass,
	groupSettingsMainInnerClass,
	groupSettingsMainScrollClass,
	groupSettingsMainScrollDirtyPadClass,
	groupSettingsSectionHeaderClass,
	groupSettingsSectionLeadClass,
	groupSettingsSectionTitleClass,
} from "@/features/group-settings/lib/group-settings-chrome";
import {
	groupSettingsProfileSchema,
	type GroupSettingsProfileFormValues,
} from "@/features/group-settings/lib/group-settings-profile-schema";
import type { GroupSettingsTarget } from "@/features/group-settings/lib/group-settings-target";
import { useReportGroupSettingsUnsaved } from "@/features/group-settings/lib/group-settings-unsaved-guard";
import { translateError } from "@/integrations/i18n/translate-error";
import { authFieldLabelClass, authInputClass } from "@/lib/form-field-styles";
import {
	settingsSectionDividerLineClass,
	settingsSectionDividerWrapClass,
	settingsSectionGapClass,
} from "@/lib/settings-chrome";
import { useQuery } from "@tanstack/react-query";
import { hasGroupPermission } from "@/lib/group-permissions";
import { cn } from "@/lib/utils";
import { myGroupsQueryOpts } from "@/services/group/group.query";
import { useSaveGroupProfileMutation } from "@/services/group/group.mutation";
import { useGroupOnlineCount } from "@/features/presence/use-group-online-count";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";

export const GROUP_SETTINGS_PROFILE_FORM_ID = "group-settings-profile-form";

const DEFAULT_BANNER_COLOR = "#313338";

const LEGACY_PRESET_TO_HEX: Record<string, string> = {
	blossom: "#FF4CD2",
	volcano: "#F23F43",
	sunset: "#FF7518",
	gold: "#FEE75C",
	grape: "#9C84EF",
	sky: "#00A8FC",
	forest: "#3BA55C",
	onyx: "#23272A",
};

function resolveGroupBannerHex(raw?: string): string {
	const trimmed = raw?.trim() ?? "";
	if (!trimmed || trimmed.toLowerCase() === "default") {
		return DEFAULT_BANNER_COLOR;
	}
	if (/^#[0-9a-fA-F]{6}$/.test(trimmed)) {
		return trimmed.toUpperCase();
	}
	return LEGACY_PRESET_TO_HEX[trimmed.toLowerCase()] ?? DEFAULT_BANNER_COLOR;
}

function groupToTargetPatch(group: Group): Partial<GroupSettingsTarget> {
	return {
		groupName: group.name,
		logoUrl: group.logo_url,
		description: group.description,
		isPrivate: group.is_private,
		bannerColor: group.banner_color,
		bannerUrl: group.banner_url,
	};
}

const descriptionInputClass = cn(authInputClass, "min-h-[6.5rem] resize-y py-2.5 sm:min-h-[7rem]");

type GroupSettingsProfileSectionProps = {
	target: GroupSettingsTarget;
};

export function GroupSettingsProfileSection({ target }: GroupSettingsProfileSectionProps) {
	const targetSyncKey = [
		target.groupId,
		target.groupName,
		target.description ?? "",
		String(target.isPrivate ?? false),
		target.bannerColor ?? "",
		target.bannerUrl ?? "",
		target.logoUrl ?? "",
	].join("\u0000");

	return <GroupSettingsProfileForm key={targetSyncKey} target={target} />;
}

function GroupSettingsProfileForm({ target }: GroupSettingsProfileSectionProps) {
	const { t } = useTranslation();
	const { patchGroupSettingsTarget } = useGroupSettings();
	const saveProfile = useSaveGroupProfileMutation();
	const { data: myGroups = [] } = useQuery(myGroupsQueryOpts());
	const membersQuery = useQuery({
		...groupMembersQueryOpts(target.groupId, ""),
		enabled: Boolean(target.groupId),
	});
	const previewMemberCount = membersQuery.data
		? (membersQuery.data.total ?? membersQuery.data.members.length)
		: undefined;
	const previewOnlineCount = useGroupOnlineCount(target.groupId);
	const currentGroup = myGroups.find((g) => g.id === target.groupId);
	const canManageGroup =
		target.role === "owner" ||
		hasGroupPermission(
			currentGroup ?? { role: target.role ?? "member", effective_permissions: [] },
			"manage_group",
		);
	const resolvedInitialBanner = resolveGroupBannerHex(target.bannerColor);

	const [iconFile, setIconFile] = useState<File | null>(null);
	const [existingLogoCleared, setExistingLogoCleared] = useState(false);
	const [bannerColor, setBannerColor] = useState(resolvedInitialBanner);

	const form = useForm<GroupSettingsProfileFormValues>({
		resolver: zodResolver(groupSettingsProfileSchema),
		defaultValues: {
			name: target.groupName,
			description: target.description ?? "",
			is_private: target.isPrivate ?? false,
		},
		mode: "onChange",
	});

	const watchedName = useWatch({ control: form.control, name: "name" });
	const watchedDescription = useWatch({ control: form.control, name: "description" });
	const dirty =
		form.formState.isDirty ||
		iconFile !== null ||
		existingLogoCleared ||
		bannerColor !== resolvedInitialBanner;

	useReportGroupSettingsUnsaved(dirty);

	const applySavedGroup = (group: Group) => {
		patchGroupSettingsTarget(groupToTargetPatch(group));
		form.reset({
			name: group.name,
			description: group.description,
			is_private: group.is_private,
		});
		setIconFile(null);
		setExistingLogoCleared(false);
		setBannerColor(resolveGroupBannerHex(group.banner_color));
	};

	const onSubmit = async (values: GroupSettingsProfileFormValues) => {
		const clear_logo = existingLogoCleared && !iconFile;
		const banner_color = bannerColor === DEFAULT_BANNER_COLOR ? "" : bannerColor;

		try {
			const group = await saveProfile.mutateAsync({
				groupId: target.groupId,
				input: {
					...values,
					banner_color,
					clear_logo,
					clear_banner_image: Boolean(target.bannerUrl?.trim()),
					iconFile,
				},
			});
			applySavedGroup(group);
			toast.success(t("group_settings.save_success"));
		} catch (err) {
			if (err instanceof ApiError) {
				const field = typeof err.params?.field === "string" ? err.params.field : "";
				if (field === "name" || field === "description") {
					form.setError(field, { message: translateError(t, err) });
					return;
				}
			}
			toast.error(translateError(t, err));
		}
	};

	const onReset = () => {
		form.reset({
			name: target.groupName,
			description: target.description ?? "",
			is_private: target.isPrivate ?? false,
		});
		setIconFile(null);
		setExistingLogoCleared(false);
		setBannerColor(resolvedInitialBanner);
	};

	return (
		<Form {...form}>
			<div className="flex min-h-0 flex-1 flex-col">
				<div
					className={cn(
						groupSettingsMainScrollClass,
						dirty && groupSettingsMainScrollDirtyPadClass,
					)}
				>
					<div className={cn(groupSettingsMainInnerClass, "flex flex-col gap-6")}>
						<div className={groupSettingsSectionHeaderClass}>
							<h2 className={groupSettingsSectionTitleClass}>
								{t("group_settings.profile.title")}
							</h2>
							<p className={groupSettingsSectionLeadClass}>{t("group_settings.profile.lead")}</p>
						</div>

						<div className="grid gap-8 lg:grid-cols-[minmax(0,1fr)_272px] lg:items-start">
							<form
								id={GROUP_SETTINGS_PROFILE_FORM_ID}
								className="flex flex-col"
								onSubmit={form.handleSubmit(onSubmit)}
							>
								<fieldset disabled={!canManageGroup} className="contents">
									<GroupIconUpload
										value={iconFile}
										onChange={(file) => {
											setIconFile(file);
											if (file) {
												setExistingLogoCleared(false);
											}
										}}
										displayName={watchedName || target.groupName}
										logoUrl={existingLogoCleared ? "" : target.logoUrl}
										onClearExistingLogo={() => setExistingLogoCleared(true)}
									/>

									<div className={settingsSectionGapClass}>
										<TextField
											control={form.control}
											name="name"
											label={t("group.home.name")}
											placeholder={t("group.home.name_placeholder")}
											autoComplete="organization"
											required
											labelClassName={authFieldLabelClass}
											inputClassName={authInputClass}
										/>
									</div>

									<div className={settingsSectionDividerWrapClass} role="presentation">
										<div className={settingsSectionDividerLineClass} />
									</div>

									<div className="space-y-2">
										<label
											htmlFor="group-settings-banner-color"
											className={cn(authFieldLabelClass, "block")}
										>
											{t("group_settings.profile.banner_color")}
										</label>
										<HexColorField
											id="group-settings-banner-color"
											value={bannerColor}
											onChange={setBannerColor}
											defaultColor={DEFAULT_BANNER_COLOR}
											resetLabel={t("group_settings.profile.banner_color_reset")}
											pickerAriaLabel={t("group_settings.profile.banner_color")}
											inputClassName={authInputClass}
										/>
									</div>

									<div className={settingsSectionDividerWrapClass} role="presentation">
										<div className={settingsSectionDividerLineClass} />
									</div>

									<FormField
										control={form.control}
										name="description"
										render={({ field }) => (
											<FormItem>
												<FormLabel className={authFieldLabelClass}>
													<FieldLabelContent>
														{t("group_settings.profile.description_label")}
													</FieldLabelContent>
												</FormLabel>
												<FormControl>
													<textarea
														{...field}
														rows={4}
														className={descriptionInputClass}
														placeholder={t("group_settings.profile.description_placeholder")}
													/>
												</FormControl>
												<FormMessage />
											</FormItem>
										)}
									/>

									<div className={settingsSectionDividerWrapClass} role="presentation">
										<div className={settingsSectionDividerLineClass} />
									</div>

									<FormField
										control={form.control}
										name="is_private"
										render={({ field }) => (
											<FormItem className="flex-row items-center justify-between gap-8 sm:gap-10">
												<div className="flex min-w-0 flex-1 flex-col gap-1">
													<FormLabel
														className={cn(authFieldLabelClass, "cursor-default capitalize")}
													>
														<FieldLabelContent>
															{t("group_settings.profile.is_private_label")}
														</FieldLabelContent>
													</FormLabel>
													<p className="text-xs leading-relaxed text-muted-foreground">
														{t("group_settings.profile.is_private_hint")}
													</p>
												</div>
												<FormControl className="shrink-0">
													<SettingsToggle
														checked={field.value}
														onCheckedChange={field.onChange}
														aria-label={t("group_settings.profile.is_private_label")}
													/>
												</FormControl>
											</FormItem>
										)}
									/>
								</fieldset>
							</form>

							<div className="lg:sticky lg:top-0">
								<GroupProfileCardPreview
									groupName={watchedName || target.groupName}
									description={watchedDescription ?? ""}
									bannerColor={bannerColor}
									logoUrl={existingLogoCleared ? "" : target.logoUrl}
									iconFile={iconFile}
									memberCount={previewMemberCount}
									onlineCount={previewOnlineCount}
								/>
							</div>
						</div>
					</div>
				</div>

				{dirty && canManageGroup ? (
					<div className={groupSettingsFormFooterFloatShellClass}>
						<div className={groupSettingsFormFooterFloatAlignClass}>
							<div className="hidden lg:block" aria-hidden />
							<div className={groupSettingsFormFooterFloatColumnClass}>
								<div className={groupSettingsMainInnerClass}>
									<GroupSettingsFormFooter
										formId={GROUP_SETTINGS_PROFILE_FORM_ID}
										errors={form.formState.errors}
										onReset={onReset}
										isPending={saveProfile.isPending}
									/>
								</div>
							</div>
						</div>
					</div>
				) : null}
			</div>
		</Form>
	);
}
