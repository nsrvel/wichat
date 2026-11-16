"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import { useEffect } from "react";
import { useForm, type FieldErrors } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { z } from "zod";
import { TextField } from "@/components/form/text-field";
import { Form } from "@/components/ui/form";
import type { CategorySettingsTarget } from "@/features/channel-category/category-settings/category-settings-sidebar";
import { GroupSettingsFormFooter } from "@/features/group-settings/components/group-settings-form-footer";
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
import type { GroupSettingsProfileFormValues } from "@/features/group-settings/lib/group-settings-profile-schema";
import { useReportGroupSettingsUnsaved } from "@/features/group-settings/lib/group-settings-unsaved-guard";
import { translateError } from "@/integrations/i18n/translate-error";
import { authFieldLabelClass, authInputClass } from "@/lib/form-field-styles";
import { cn } from "@/lib/utils";
import { useUpdateGroupCategoryMutation } from "@/services/channel-category/channel-category.mutation";
import * as channelQuery from "@/services/channel/channel.query";

const CATEGORY_OVERVIEW_FORM_ID = "category-settings-overview-form";

const CategoryOverviewSchema = z.object({
	name: z
		.string()
		.trim()
		.min(1, "channel_category.settings.category_name_required")
		.max(64, "channel_category.settings.category_name_max"),
});

type CategoryOverviewFormValues = z.infer<typeof CategoryOverviewSchema>;

export function CategorySettingsOverviewSection({ target }: { target: CategorySettingsTarget }) {
	const { t } = useTranslation();
	const updateMutation = useUpdateGroupCategoryMutation(target.groupId);
	const channelsQuery = useQuery(channelQuery.groupChannelsQueryOpts(target.groupId));
	const liveCategory = channelsQuery.data?.categories.find((c) => c.id === target.categoryId);
	const currentName = liveCategory?.name ?? target.categoryName;

	const form = useForm<CategoryOverviewFormValues>({
		resolver: zodResolver(CategoryOverviewSchema),
		defaultValues: { name: currentName },
		mode: "onChange",
	});

	const { isDirty } = form.formState;
	useReportGroupSettingsUnsaved(isDirty);

	useEffect(() => {
		form.reset({ name: currentName });
	}, [currentName, form]);

	const handleSubmit = form.handleSubmit((values) => {
		updateMutation.mutate(
			{
				categoryId: target.categoryId,
				payload: { name: values.name },
			},
			{
				onSuccess: (updated) => {
					form.reset({ name: updated.name });
					toast.success(t("channel_category.settings.updated_toast"));
				},
				onError: (err) => {
					toast.error(translateError(t, err));
				},
			},
		);
	});

	const handleReset = () => {
		form.reset({ name: currentName });
	};

	return (
		<div className="flex min-h-0 flex-1 flex-col">
			<div
				className={cn(
					groupSettingsMainScrollClass,
					isDirty && groupSettingsMainScrollDirtyPadClass,
				)}
			>
				<div className={groupSettingsMainInnerClass}>
					<header className={groupSettingsSectionHeaderClass}>
						<h2 className={groupSettingsSectionTitleClass}>
							{t("channel_category.settings.overview")}
						</h2>
						<p className={groupSettingsSectionLeadClass}>
							{t("channel_category.settings.overview_subtitle")}
						</p>
					</header>

					<Form {...form}>
						<form id={CATEGORY_OVERVIEW_FORM_ID} onSubmit={handleSubmit} className="max-w-lg">
							<TextField
								control={form.control}
								name="name"
								label={t("channel_category.settings.category_name_label")}
								placeholder={t("channel_category.settings.category_name_placeholder")}
								required
								maxLength={64}
								labelClassName={authFieldLabelClass}
								inputClassName={authInputClass}
							/>
						</form>
					</Form>
				</div>
			</div>

			{isDirty ? (
				<div className={groupSettingsFormFooterFloatShellClass}>
					<div className={groupSettingsFormFooterFloatAlignClass}>
						<div className={groupSettingsFormFooterFloatColumnClass}>
							<GroupSettingsFormFooter
								formId={CATEGORY_OVERVIEW_FORM_ID}
								errors={form.formState.errors as FieldErrors<GroupSettingsProfileFormValues>}
								onReset={handleReset}
								isPending={updateMutation.isPending}
							/>
						</div>
					</div>
				</div>
			) : null}
		</div>
	);
}
