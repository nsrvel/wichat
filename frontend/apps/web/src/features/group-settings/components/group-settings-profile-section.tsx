"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { FieldLabelContent } from "@/components/form/field-label";
import { TextField } from "@/components/form/text-field";
import {
	Form,
	FormControl,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { GroupSettingsFormFooter } from "@/features/group-settings/components/group-settings-form-footer";
import { GroupSettingsToggle } from "@/features/group-settings/components/group-settings-toggle";
import { authFieldLabelClass, authInputClass } from "@/features/auth/components/auth-shell";
import { GroupBannerField } from "@/features/group/components/group-banner-field";
import { GroupIconUpload } from "@/features/group/components/group-icon-upload";
import {
	GROUP_BANNER_COLOR_DEFAULT,
	GROUP_BANNER_IMAGE_ID,
	isGroupBannerImageId,
	normalizeGroupBannerColor,
	type GroupBannerColorId,
} from "@/features/group/lib/group-banner-presets";
import { useGroupSettings } from "@/features/group-settings/group-settings-provider";
import {
	groupSettingsFormFooterFloatAlignClass,
	groupSettingsFormFooterFloatColumnClass,
	groupSettingsFormFooterFloatShellClass,
	groupSettingsMainInnerClass,
	groupSettingsMainScrollClass,
	groupSettingsMainScrollDirtyPadClass,
	groupSettingsProfileSectionDividerLineClass,
	groupSettingsProfileSectionDividerWrapClass,
	groupSettingsProfileSectionGapClass,
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
import { cn } from "@/lib/utils";
import { useSaveGroupProfileMutation } from "@/services/group/group.mutation";
import type { Group } from "@wichat/contracts";

export const GROUP_SETTINGS_PROFILE_FORM_ID = "group-settings-profile-form";

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

function bannerSelectionFromTarget(target: GroupSettingsTarget): GroupBannerColorId {
	return target.bannerUrl?.trim()
		? GROUP_BANNER_IMAGE_ID
		: normalizeGroupBannerColor(target.bannerColor);
}

const descriptionInputClass = cn(authInputClass, "min-h-[6.5rem] resize-y py-2.5 sm:min-h-[7rem]");

type GroupSettingsProfileSectionProps = {
	target: GroupSettingsTarget;
};

export function GroupSettingsProfileSection({ target }: GroupSettingsProfileSectionProps) {
	const { t } = useTranslation();
	const { patchGroupSettingsTarget } = useGroupSettings();
	const saveProfile = useSaveGroupProfileMutation();
	const [iconFile, setIconFile] = useState<File | null>(null);
	const [existingLogoCleared, setExistingLogoCleared] = useState(false);
	const [bannerSelection, setBannerSelection] = useState<GroupBannerColorId>(() =>
		target.bannerUrl?.trim()
			? GROUP_BANNER_IMAGE_ID
			: normalizeGroupBannerColor(target.bannerColor),
	);
	const [bannerFile, setBannerFile] = useState<File | null>(null);
	const [existingBannerCleared, setExistingBannerCleared] = useState(false);

	const form = useForm<GroupSettingsProfileFormValues>({
		resolver: zodResolver(groupSettingsProfileSchema),
		defaultValues: {
			name: target.groupName,
			description: target.description ?? "",
			is_private: target.isPrivate ?? false,
		},
		mode: "onChange",
	});

	useEffect(() => {
		form.reset({
			name: target.groupName,
			description: target.description ?? "",
			is_private: target.isPrivate ?? false,
		});
		// eslint-disable-next-line react-hooks/set-state-in-effect -- sync form & local asset states on target switch
		setIconFile(null);
		setExistingLogoCleared(false);
		setBannerSelection(
			target.bannerUrl?.trim()
				? GROUP_BANNER_IMAGE_ID
				: normalizeGroupBannerColor(target.bannerColor),
		);
		setBannerFile(null);
		setExistingBannerCleared(false);
	}, [
		form,
		target.bannerColor,
		target.bannerUrl,
		target.description,
		target.groupId,
		target.groupName,
		target.isPrivate,
	]);

	const watchedName = useWatch({ control: form.control, name: "name" });
	const savedBannerSelection: GroupBannerColorId = target.bannerUrl?.trim()
		? GROUP_BANNER_IMAGE_ID
		: normalizeGroupBannerColor(target.bannerColor);
	const bannerDirty =
		bannerSelection !== savedBannerSelection || bannerFile !== null || existingBannerCleared;
	const dirty = form.formState.isDirty || iconFile !== null || existingLogoCleared || bannerDirty;

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
		setBannerSelection(
			bannerSelectionFromTarget({
				...target,
				...groupToTargetPatch(group),
			}),
		);
		setBannerFile(null);
		setExistingBannerCleared(false);
	};

	const onSubmit = async (values: GroupSettingsProfileFormValues) => {
		const isImageMode = isGroupBannerImageId(bannerSelection);
		const savedHadImage = Boolean(target.bannerUrl?.trim());
		const clear_logo = existingLogoCleared && !iconFile;
		const clear_banner_image = existingBannerCleared || (savedHadImage && !isImageMode);
		const banner_color = isImageMode
			? GROUP_BANNER_COLOR_DEFAULT
			: bannerSelection === GROUP_BANNER_COLOR_DEFAULT
				? GROUP_BANNER_COLOR_DEFAULT
				: bannerSelection;

		try {
			const group = await saveProfile.mutateAsync({
				groupId: target.groupId,
				input: {
					...values,
					banner_color,
					clear_logo,
					clear_banner_image,
					iconFile,
					bannerFile,
				},
			});
			applySavedGroup(group);
			toast.success(t("group_settings.save_success"));
		} catch (err) {
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
		setBannerSelection(savedBannerSelection);
		setBannerFile(null);
		setExistingBannerCleared(false);
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

						<form
							id={GROUP_SETTINGS_PROFILE_FORM_ID}
							className="flex flex-col"
							onSubmit={form.handleSubmit(onSubmit)}
						>
							<GroupIconUpload
								value={iconFile}
								onChange={setIconFile}
								displayName={watchedName || target.groupName}
								logoUrl={existingLogoCleared ? "" : target.logoUrl}
								onClearExistingLogo={() => setExistingLogoCleared(true)}
							/>

							<div className={groupSettingsProfileSectionGapClass}>
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

							<div className={groupSettingsProfileSectionDividerWrapClass} role="presentation">
								<div className={groupSettingsProfileSectionDividerLineClass} />
							</div>

							<GroupBannerField
								selection={bannerSelection}
								onSelectionChange={setBannerSelection}
								imageFile={bannerFile}
								onImageFileChange={setBannerFile}
								existingImageUrl={
									isGroupBannerImageId(bannerSelection) && !existingBannerCleared
										? target.bannerUrl
										: ""
								}
								onClearExistingImage={() => setExistingBannerCleared(true)}
							/>

							<div className={groupSettingsProfileSectionDividerWrapClass} role="presentation">
								<div className={groupSettingsProfileSectionDividerLineClass} />
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

							<div className={groupSettingsProfileSectionDividerWrapClass} role="presentation">
								<div className={groupSettingsProfileSectionDividerLineClass} />
							</div>

							<FormField
								control={form.control}
								name="is_private"
								render={({ field }) => (
									<FormItem className="flex-row items-center justify-between gap-8 sm:gap-10">
										<div className="flex min-w-0 flex-1 flex-col gap-1">
											<FormLabel className={cn(authFieldLabelClass, "cursor-default capitalize")}>
												<FieldLabelContent>
													{t("group_settings.profile.is_private_label")}
												</FieldLabelContent>
											</FormLabel>
											<p className="text-xs leading-relaxed text-muted-foreground">
												{t("group_settings.profile.is_private_hint")}
											</p>
										</div>
										<FormControl className="shrink-0">
											<GroupSettingsToggle
												checked={field.value}
												onCheckedChange={field.onChange}
												aria-label={t("group_settings.profile.is_private_label")}
											/>
										</FormControl>
									</FormItem>
								)}
							/>
						</form>
					</div>
				</div>

				{dirty ? (
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
