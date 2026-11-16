"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import type {
	ChannelContentVisibility,
	ForumDefaultLayout,
	ForumSortOrder,
	ForumTagMatching,
	VoiceVideoQuality,
} from "@wichat/contracts/chat/channel";
import { useEffect, useMemo } from "react";
import { useForm, useWatch, type FieldErrors } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { z } from "zod";
import { TextField } from "@/components/form/text-field";
import { CustomSelect, type SelectOption } from "@/components/ui/custom-select";
import { Form } from "@/components/ui/form";
import type { ChannelSettingsTarget } from "@/features/channel/channel-settings/channel-settings-sidebar";
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
import { useUpdateGroupChannelMutation } from "@/services/channel/channel.mutation";
import * as channelQuery from "@/services/channel/channel.query";

const CHANNEL_OVERVIEW_FORM_ID = "channel-settings-overview-form";

const SLOWMODE_OPTIONS: Array<{ value: number; labelKey: string }> = [
	{ value: 0, labelKey: "channel.channel_settings.slowmode_off" },
	{ value: 5, labelKey: "channel.channel_settings.slowmode_5s" },
	{ value: 10, labelKey: "channel.channel_settings.slowmode_10s" },
	{ value: 15, labelKey: "channel.channel_settings.slowmode_15s" },
	{ value: 30, labelKey: "channel.channel_settings.slowmode_30s" },
	{ value: 60, labelKey: "channel.channel_settings.slowmode_1m" },
	{ value: 120, labelKey: "channel.channel_settings.slowmode_2m" },
	{ value: 300, labelKey: "channel.channel_settings.slowmode_5m" },
	{ value: 600, labelKey: "channel.channel_settings.slowmode_10m" },
	{ value: 900, labelKey: "channel.channel_settings.slowmode_15m" },
	{ value: 1800, labelKey: "channel.channel_settings.slowmode_30m" },
	{ value: 3600, labelKey: "channel.channel_settings.slowmode_1h" },
	{ value: 7200, labelKey: "channel.channel_settings.slowmode_2h" },
	{ value: 21600, labelKey: "channel.channel_settings.slowmode_6h" },
];

const RTC_REGION_OPTIONS: Array<{ value: string; labelKey: string }> = [
	{ value: "automatic", labelKey: "channel.channel_settings.rtc_region_automatic" },
	{ value: "singapore", labelKey: "channel.channel_settings.rtc_region_singapore" },
	{ value: "japan", labelKey: "channel.channel_settings.rtc_region_japan" },
	{ value: "us-west", labelKey: "channel.channel_settings.rtc_region_us_west" },
	{ value: "us-east", labelKey: "channel.channel_settings.rtc_region_us_east" },
	{ value: "europe", labelKey: "channel.channel_settings.rtc_region_europe" },
	{ value: "sydney", labelKey: "channel.channel_settings.rtc_region_sydney" },
];

const ChannelOverviewSchema = z.object({
	name: z
		.string()
		.trim()
		.min(1, "channel.channel_settings.channel_name_required")
		.max(64, "channel.channel_settings.channel_name_max"),
	slowmode_seconds: z.number().int().min(0).max(21600),
	thread_slowmode_seconds: z.number().int().min(0).max(21600),
	content_visibility: z.enum(["default", "spoiler", "age_restricted"]),
	forum_default_layout: z.enum(["list", "gallery"]),
	forum_sort_order: z.enum(["recent_activity", "creation_time"]),
	forum_tag_matching: z.enum(["match_some", "match_all"]),
	bitrate_kbps: z.number().int().min(8).max(96),
	video_quality: z.enum(["auto", "720p"]),
	user_limit: z.number().int().min(0).max(99),
	rtc_region: z.string().min(1),
});

type ChannelOverviewFormValues = z.infer<typeof ChannelOverviewSchema>;

const customSelectTriggerClass =
	"flex h-11 w-full items-center justify-between rounded-xl border border-input bg-field px-3.5 text-sm font-medium text-foreground transition-colors hover:border-foreground/30 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40";

export function ChannelSettingsOverviewSection({ target }: { target: ChannelSettingsTarget }) {
	const { t } = useTranslation();
	const updateMutation = useUpdateGroupChannelMutation(target.groupId);
	const channelsQuery = useQuery(channelQuery.groupChannelsQueryOpts(target.groupId));
	const liveChannel = channelsQuery.data?.channels.find((c) => c.id === target.channelId);

	const channelType = liveChannel?.type ?? target.channelType;

	const initialValues = useMemo<ChannelOverviewFormValues>(
		() => ({
			name: liveChannel?.name ?? target.channelName,
			slowmode_seconds: liveChannel?.slowmode_seconds ?? 0,
			thread_slowmode_seconds: liveChannel?.thread_slowmode_seconds ?? 0,
			content_visibility: (liveChannel?.content_visibility ??
				"default") as ChannelContentVisibility,
			forum_default_layout: (liveChannel?.forum_default_layout ?? "list") as ForumDefaultLayout,
			forum_sort_order: (liveChannel?.forum_sort_order ?? "recent_activity") as ForumSortOrder,
			forum_tag_matching: (liveChannel?.forum_tag_matching ?? "match_some") as ForumTagMatching,
			bitrate_kbps: liveChannel?.bitrate_kbps ?? 64,
			video_quality: (liveChannel?.video_quality ?? "auto") as VoiceVideoQuality,
			user_limit: liveChannel?.user_limit ?? 0,
			rtc_region: liveChannel?.rtc_region ?? "automatic",
		}),
		[liveChannel, target.channelName],
	);

	const form = useForm<ChannelOverviewFormValues>({
		resolver: zodResolver(ChannelOverviewSchema),
		defaultValues: initialValues,
		mode: "onChange",
	});

	const { isDirty } = form.formState;
	useReportGroupSettingsUnsaved(isDirty);

	useEffect(() => {
		form.reset(initialValues);
	}, [form, initialValues]);

	const handleSubmit = form.handleSubmit((values) => {
		updateMutation.mutate(
			{
				channelId: target.channelId,
				payload: {
					name: values.name,
					slowmode_seconds: values.slowmode_seconds,
					thread_slowmode_seconds: values.thread_slowmode_seconds,
					content_visibility: values.content_visibility,
					forum_default_layout: values.forum_default_layout,
					forum_sort_order: values.forum_sort_order,
					forum_tag_matching: values.forum_tag_matching,
					bitrate_kbps: values.bitrate_kbps,
					video_quality: values.video_quality,
					user_limit: values.user_limit,
					rtc_region: values.rtc_region,
				},
			},
			{
				onSuccess: (updated) => {
					form.reset({
						name: updated.name,
						slowmode_seconds: updated.slowmode_seconds,
						thread_slowmode_seconds: updated.thread_slowmode_seconds,
						content_visibility: updated.content_visibility,
						forum_default_layout: updated.forum_default_layout,
						forum_sort_order: updated.forum_sort_order,
						forum_tag_matching: updated.forum_tag_matching,
						bitrate_kbps: updated.bitrate_kbps,
						video_quality: updated.video_quality,
						user_limit: updated.user_limit,
						rtc_region: updated.rtc_region,
					});
					toast.success(t("channel.channel_settings.updated_toast"));
				},
				onError: (err) => {
					toast.error(translateError(t, err));
				},
			},
		);
	});

	const handleReset = () => {
		form.reset(initialValues);
	};

	const watched = useWatch({ control: form.control });
	const visibilityValue = watched.content_visibility ?? "default";
	const videoQualityValue = watched.video_quality ?? "auto";
	const bitrateValue = watched.bitrate_kbps ?? 64;
	const userLimitValue = watched.user_limit ?? 0;
	const slowmodeValue = watched.slowmode_seconds ?? 0;
	const threadSlowmodeValue = watched.thread_slowmode_seconds ?? 0;
	const forumLayoutValue = watched.forum_default_layout ?? "list";
	const forumSortOrderValue = watched.forum_sort_order ?? "recent_activity";
	const forumTagMatchingValue = watched.forum_tag_matching ?? "match_some";
	const rtcRegionValue = watched.rtc_region ?? "automatic";

	const slowmodeSelectOptions: SelectOption[] = useMemo(
		() => SLOWMODE_OPTIONS.map((opt) => ({ value: String(opt.value), label: t(opt.labelKey) })),
		[t],
	);

	const rtcRegionSelectOptions: SelectOption[] = useMemo(
		() => RTC_REGION_OPTIONS.map((opt) => ({ value: opt.value, label: t(opt.labelKey) })),
		[t],
	);

	const forumLayoutSelectOptions: SelectOption[] = useMemo(
		() => [
			{ value: "list", label: t("channel.channel_settings.forum_layout_list") },
			{ value: "gallery", label: t("channel.channel_settings.forum_layout_gallery") },
		],
		[t],
	);

	const forumSortOrderSelectOptions: SelectOption[] = useMemo(
		() => [
			{ value: "recent_activity", label: t("channel.channel_settings.forum_sort_recent_activity") },
			{ value: "creation_time", label: t("channel.channel_settings.forum_sort_creation_time") },
		],
		[t],
	);

	const forumTagMatchingSelectOptions: SelectOption[] = useMemo(
		() => [
			{ value: "match_some", label: t("channel.channel_settings.forum_tag_match_some") },
			{ value: "match_all", label: t("channel.channel_settings.forum_tag_match_all") },
		],
		[t],
	);

	const visibilityOptions: Array<{
		value: ChannelContentVisibility;
		titleKey: string;
		descKey: string;
	}> = [
		{
			value: "default",
			titleKey: "channel.channel_settings.visibility_default_title",
			descKey: "channel.channel_settings.visibility_default_desc",
		},
		{
			value: "spoiler",
			titleKey: "channel.channel_settings.visibility_spoiler_title",
			descKey: "channel.channel_settings.visibility_spoiler_desc",
		},
		{
			value: "age_restricted",
			titleKey: "channel.channel_settings.visibility_age_title",
			descKey: "channel.channel_settings.visibility_age_desc",
		},
	];

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
							{t("channel.channel_settings.overview")}
						</h2>
						<p className={groupSettingsSectionLeadClass}>
							{t("channel.channel_settings.overview_subtitle")}
						</p>
					</header>

					<Form {...form}>
						<form
							id={CHANNEL_OVERVIEW_FORM_ID}
							onSubmit={handleSubmit}
							className="max-w-lg space-y-6"
						>
							<TextField
								control={form.control}
								name="name"
								label={t("channel.channel_settings.channel_name_label")}
								placeholder={t("channel.channel_settings.channel_name_placeholder")}
								required
								maxLength={64}
								labelClassName={authFieldLabelClass}
								inputClassName={authInputClass}
							/>

							{channelType === "text_only" || channelType === "voice_only" ? (
								<div className="space-y-1.5 border-t border-border/60 pt-5">
									<span className={authFieldLabelClass}>
										{t("channel.channel_settings.slowmode_label")}
									</span>
									<CustomSelect
										options={slowmodeSelectOptions}
										value={String(slowmodeValue)}
										placeholder={t("channel.channel_settings.slowmode_off")}
										onChange={(val) =>
											form.setValue("slowmode_seconds", Number(val), { shouldDirty: true })
										}
										portaled
										triggerClassName={customSelectTriggerClass}
										menuClassName="z-[90]"
									/>
									<p className="text-xs text-muted-foreground">
										{t("channel.channel_settings.slowmode_description")}
									</p>
								</div>
							) : null}

							{channelType === "forum" ? (
								<>
									<div className="space-y-4 border-t border-border/60 pt-5">
										<h3 className="text-sm font-semibold text-foreground">
											{t("channel.channel_settings.slowmode_label")}
										</h3>
										<div className="space-y-1.5">
											<span className={authFieldLabelClass}>
												{t("channel.channel_settings.slowmode_posts_label")}
											</span>
											<CustomSelect
												options={slowmodeSelectOptions}
												value={String(slowmodeValue)}
												placeholder={t("channel.channel_settings.slowmode_off")}
												onChange={(val) =>
													form.setValue("slowmode_seconds", Number(val), { shouldDirty: true })
												}
												portaled
												triggerClassName={customSelectTriggerClass}
												menuClassName="z-[90]"
											/>
											<p className="text-xs text-muted-foreground">
												{t("channel.channel_settings.slowmode_posts_description")}
											</p>
										</div>

										<div className="space-y-1.5">
											<span className={authFieldLabelClass}>
												{t("channel.channel_settings.slowmode_messages_label")}
											</span>
											<CustomSelect
												options={slowmodeSelectOptions}
												value={String(threadSlowmodeValue)}
												placeholder={t("channel.channel_settings.slowmode_off")}
												onChange={(val) =>
													form.setValue("thread_slowmode_seconds", Number(val), {
														shouldDirty: true,
													})
												}
												portaled
												triggerClassName={customSelectTriggerClass}
												menuClassName="z-[90]"
											/>
											<p className="text-xs text-muted-foreground">
												{t("channel.channel_settings.slowmode_messages_description")}
											</p>
										</div>
									</div>

									<div className="space-y-4 border-t border-border/60 pt-5">
										<div className="space-y-1.5">
											<span className={authFieldLabelClass}>
												{t("channel.channel_settings.forum_default_layout_label")}
											</span>
											<p className="text-xs text-muted-foreground">
												{t("channel.channel_settings.forum_default_layout_desc")}
											</p>
											<CustomSelect
												options={forumLayoutSelectOptions}
												value={forumLayoutValue}
												placeholder={t("channel.channel_settings.forum_layout_list")}
												onChange={(val) =>
													form.setValue("forum_default_layout", val as ForumDefaultLayout, {
														shouldDirty: true,
													})
												}
												portaled
												triggerClassName={customSelectTriggerClass}
												menuClassName="z-[90]"
											/>
										</div>

										<div className="space-y-1.5">
											<span className={authFieldLabelClass}>
												{t("channel.channel_settings.forum_sort_order_label")}
											</span>
											<p className="text-xs text-muted-foreground">
												{t("channel.channel_settings.forum_sort_order_desc")}
											</p>
											<CustomSelect
												options={forumSortOrderSelectOptions}
												value={forumSortOrderValue}
												placeholder={t("channel.channel_settings.forum_sort_recent_activity")}
												onChange={(val) =>
													form.setValue("forum_sort_order", val as ForumSortOrder, {
														shouldDirty: true,
													})
												}
												portaled
												triggerClassName={customSelectTriggerClass}
												menuClassName="z-[90]"
											/>
										</div>

										<div className="space-y-1.5">
											<span className={authFieldLabelClass}>
												{t("channel.channel_settings.forum_tag_matching_label")}
											</span>
											<p className="text-xs text-muted-foreground">
												{t("channel.channel_settings.forum_tag_matching_desc")}
											</p>
											<CustomSelect
												options={forumTagMatchingSelectOptions}
												value={forumTagMatchingValue}
												placeholder={t("channel.channel_settings.forum_tag_match_some")}
												onChange={(val) =>
													form.setValue("forum_tag_matching", val as ForumTagMatching, {
														shouldDirty: true,
													})
												}
												portaled
												triggerClassName={customSelectTriggerClass}
												menuClassName="z-[90]"
											/>
										</div>
									</div>
								</>
							) : null}

							<div className="space-y-1.5 border-t border-border/60 pt-5">
								<span className={authFieldLabelClass}>
									{t("channel.channel_settings.content_visibility_label")}
								</span>
								<div
									className="space-y-1"
									role="radiogroup"
									aria-label={t("channel.channel_settings.content_visibility_label")}
								>
									{visibilityOptions.map((opt) => {
										const selected = visibilityValue === opt.value;
										return (
											<button
												key={opt.value}
												type="button"
												role="radio"
												aria-checked={selected}
												onClick={() =>
													form.setValue("content_visibility", opt.value, { shouldDirty: true })
												}
												className={cn(
													"flex w-full items-start gap-3 rounded-lg px-2 py-2 text-left transition-colors hover:bg-muted/40",
													selected ? "bg-muted/35" : "",
												)}
											>
												<span
													className={cn(
														"mt-0.5 flex size-[18px] shrink-0 items-center justify-center rounded-full border-2 transition-colors",
														selected ? "border-primary bg-primary" : "border-muted-foreground/50",
													)}
													aria-hidden
												>
													{selected ? (
														<span className="size-1.5 rounded-full bg-primary-foreground" />
													) : null}
												</span>
												<span className="min-w-0 flex-1 space-y-0.5">
													<span className="block text-sm font-semibold text-foreground">
														{t(opt.titleKey)}
													</span>
													<span className="block text-xs leading-snug text-muted-foreground">
														{t(opt.descKey)}
													</span>
												</span>
											</button>
										);
									})}
								</div>
							</div>

							{channelType === "voice_only" ? (
								<div className="space-y-6 border-t border-border/60 pt-5">
									<div className="space-y-2">
										<div className="flex items-center justify-between">
											<label htmlFor="voice-bitrate" className={authFieldLabelClass}>
												{t("channel.channel_settings.bitrate_label")}
											</label>
											<span className="text-xs font-medium text-foreground tabular-nums">
												{bitrateValue}kbps
											</span>
										</div>
										<input
											id="voice-bitrate"
											type="range"
											min={8}
											max={96}
											step={1}
											value={bitrateValue}
											onChange={(e) =>
												form.setValue("bitrate_kbps", Number(e.target.value), {
													shouldDirty: true,
												})
											}
											className="h-2 w-full cursor-pointer accent-primary"
										/>
										<div className="flex justify-between text-[11px] text-muted-foreground">
											<span>8kbps</span>
											<span>64kbps</span>
											<span>96kbps</span>
										</div>
										<p className="text-xs text-muted-foreground">
											{t("channel.channel_settings.bitrate_desc")}
										</p>
									</div>

									<div className="space-y-1.5">
										<span className={authFieldLabelClass}>
											{t("channel.channel_settings.video_quality_label")}
										</span>
										<div
											className="space-y-1"
											role="radiogroup"
											aria-label={t("channel.channel_settings.video_quality_label")}
										>
											{(["auto", "720p"] as const).map((qual) => {
												const selected = videoQualityValue === qual;
												return (
													<button
														key={qual}
														type="button"
														role="radio"
														aria-checked={selected}
														onClick={() =>
															form.setValue("video_quality", qual, { shouldDirty: true })
														}
														className={cn(
															"flex w-full items-center gap-3 rounded-lg px-2 py-2 text-left transition-colors hover:bg-muted/40",
															selected ? "bg-muted/35" : "",
														)}
													>
														<span
															className={cn(
																"flex size-[18px] shrink-0 items-center justify-center rounded-full border-2 transition-colors",
																selected
																	? "border-primary bg-primary"
																	: "border-muted-foreground/50",
															)}
															aria-hidden
														>
															{selected ? (
																<span className="size-1.5 rounded-full bg-primary-foreground" />
															) : null}
														</span>
														<span className="text-sm font-semibold text-foreground">
															{t(
																qual === "auto"
																	? "channel.channel_settings.video_quality_auto"
																	: "channel.channel_settings.video_quality_720p",
															)}
														</span>
													</button>
												);
											})}
										</div>
										<p className="text-xs text-muted-foreground">
											{t("channel.channel_settings.video_quality_desc")}
										</p>
									</div>

									<div className="space-y-2">
										<div className="flex items-center justify-between">
											<label htmlFor="voice-user-limit" className={authFieldLabelClass}>
												{t("channel.channel_settings.user_limit_label")}
											</label>
											<span className="text-xs font-medium text-foreground tabular-nums">
												{userLimitValue === 0
													? t("channel.channel_settings.user_limit_unlimited")
													: t("channel.channel_settings.user_limit_users", {
															count: userLimitValue,
														})}
											</span>
										</div>
										<input
											id="voice-user-limit"
											type="range"
											min={0}
											max={99}
											step={1}
											value={userLimitValue}
											onChange={(e) =>
												form.setValue("user_limit", Number(e.target.value), {
													shouldDirty: true,
												})
											}
											className="h-2 w-full cursor-pointer accent-primary"
										/>
										<div className="flex justify-between text-[11px] text-muted-foreground">
											<span>∞</span>
											<span>99</span>
										</div>
										<p className="text-xs text-muted-foreground">
											{t("channel.channel_settings.user_limit_desc")}
										</p>
									</div>

									<div className="space-y-1.5">
										<span className={authFieldLabelClass}>
											{t("channel.channel_settings.rtc_region_label")}
										</span>
										<p className="text-xs text-muted-foreground">
											{t("channel.channel_settings.rtc_region_desc")}
										</p>
										<CustomSelect
											options={rtcRegionSelectOptions}
											value={rtcRegionValue}
											placeholder={t("channel.channel_settings.rtc_region_automatic")}
											onChange={(val) => form.setValue("rtc_region", val, { shouldDirty: true })}
											portaled
											triggerClassName={customSelectTriggerClass}
											menuClassName="z-[90]"
										/>
									</div>
								</div>
							) : null}
						</form>
					</Form>
				</div>
			</div>

			{isDirty ? (
				<div className={groupSettingsFormFooterFloatShellClass}>
					<div className={groupSettingsFormFooterFloatAlignClass}>
						<div className={groupSettingsFormFooterFloatColumnClass}>
							<GroupSettingsFormFooter
								formId={CHANNEL_OVERVIEW_FORM_ID}
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
