"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft, ChevronRight } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useForm, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { TextField } from "@/components/form/text-field";
import { Button } from "@/components/ui/button";
import { Form } from "@/components/ui/form";
import {
	WiChatDialog,
	WiChatDialogBody,
	WiChatDialogFooter,
	WiChatDialogFooterActions,
	WiChatDialogHeader,
	WiChatDialogStepCarousel,
} from "@/components/ui/wichat-dialog";
import {
	authFieldLabelClass,
	authFormClass,
	authInputClass,
	authSubmitClass,
	authSubmitGhostClass,
	groupOnboardingCardShellClass,
} from "@/features/auth/components/auth-shell";
import { GroupIconUpload } from "@/features/group/components/group-icon-upload";
import { translateError } from "@/integrations/i18n/translate-error";
import { displayNameSchema, type Group } from "@wichat/contracts";
import { cn } from "@/lib/utils";
import * as groupQuery from "@/services/group/group.query";
import { channelPath } from "@/lib/app-routes";
import { Q_MY_GROUPS } from "@/services/group/group.key";
import { leaveGroup, uploadGroupIcon } from "@/services/group/group.service";
import { useCreateGroupMutation, useJoinGroupMutation } from "@/services/group/group.mutation";

const GROUP_ONBOARDING_TITLE_ID = "group-onboarding-title";
const GROUP_CREATE_FORM_ID = "wichat-group-create-form";
const authFormFooterButtonClass = "w-auto min-w-[7.5rem] max-w-[55%] shrink-0 px-8";

const createSchema = z.object({ name: displayNameSchema });

type Mode = "pick" | "create" | "join";

const GROUP_TEMPLATES = [
	{
		id: "gaming",
		emoji: "🎮",
		labelKey: "group.home.template_gaming",
		nameKey: "group.home.template_gaming_name",
	},
	{
		id: "friends",
		emoji: "💜",
		labelKey: "group.home.template_friends",
		nameKey: "group.home.template_friends_name",
	},
	{
		id: "study",
		emoji: "🍎",
		labelKey: "group.home.template_study",
		nameKey: "group.home.template_study_name",
	},
	{
		id: "school",
		emoji: "📖",
		labelKey: "group.home.template_school",
		nameKey: "group.home.template_school_name",
	},
] as const;

const optionRowClass =
	"flex w-full min-w-0 items-center gap-3 rounded-xl bg-field px-3 py-2.5 text-left text-sm font-medium text-foreground transition-colors hover:bg-field/80 sm:py-3";

const optionEmojiClass =
	"flex h-10 w-10 shrink-0 items-center justify-center text-[1.625rem] leading-none";

function modeStepIndex(mode: Mode): number {
	if (mode === "pick") {
		return 0;
	}
	if (mode === "create") {
		return 1;
	}

	return 2;
}

type GroupOnboardingFlowProps = {
	onComplete?: () => void;
	presentation?: "dialog" | "page";
};

export function GroupOnboardingFlow({
	onComplete,
	presentation = "page",
}: GroupOnboardingFlowProps) {
	const { t } = useTranslation();
	const [mode, setMode] = useState<Mode>("pick");
	const stepIndex = modeStepIndex(mode);
	const joinable = useQuery({ ...groupQuery.joinableGroupsQueryOpts(), enabled: mode === "join" });
	const router = useRouter();
	const queryClient = useQueryClient();
	const createGroup = useCreateGroupMutation();
	const joinGroup = useJoinGroupMutation();
	const form = useForm<z.infer<typeof createSchema>>({
		resolver: zodResolver(createSchema),
		defaultValues: { name: "" },
	});
	const [groupIconFile, setGroupIconFile] = useState<File | null>(null);
	const [createBusy, setCreateBusy] = useState(false);
	const watchedName = useWatch({ control: form.control, name: "name" });

	const finish = () => {
		onComplete?.();
		setMode("pick");
		form.reset();
		setGroupIconFile(null);
	};

	const openGroupChannel = (groupId: string, defaultChannelId: string) => {
		finish();
		router.push(channelPath(groupId, defaultChannelId));
	};

	const startCreate = (prefillName = "") => {
		form.reset({ name: prefillName });
		setGroupIconFile(null);
		setMode("create");
	};

	const goJoin = () => {
		setMode("join");
	};

	const backToPick = () => {
		setGroupIconFile(null);
		setMode("pick");
	};

	const pickChrome = (
		<>
			<WiChatDialogHeader
				title={t("group.home.modal_title")}
				subtitle={t("group.home.modal_lead")}
				titleId={stepIndex === 0 ? GROUP_ONBOARDING_TITLE_ID : undefined}
			/>
			<WiChatDialogBody>
				<div className="space-y-5 pb-4 sm:space-y-6 sm:pb-5">
					<button type="button" className={optionRowClass} onClick={() => startCreate()}>
						<span className={optionEmojiClass} aria-hidden>
							🏆
						</span>
						<span className="min-w-0 flex-1">{t("group.home.create_own")}</span>
						<ChevronRight className="h-5 w-5 shrink-0 text-muted-foreground" aria-hidden />
					</button>

					<div className="space-y-2">
						<p className="text-[11px] font-bold tracking-wide text-muted-foreground uppercase">
							{t("group.home.templates_section")}
						</p>
						<ul className="space-y-1.5">
							{GROUP_TEMPLATES.map((template) => (
								<li key={template.id}>
									<button
										type="button"
										className={optionRowClass}
										onClick={() => startCreate(t(template.nameKey))}
									>
										<span className={optionEmojiClass} aria-hidden>
											{template.emoji}
										</span>
										<span className="min-w-0 flex-1">{t(template.labelKey)}</span>
										<ChevronRight className="h-5 w-5 shrink-0 text-muted-foreground" aria-hidden />
									</button>
								</li>
							))}
						</ul>
					</div>
				</div>
			</WiChatDialogBody>
			<WiChatDialogFooter>
				<div className="space-y-3 text-center">
					<p className="text-sm font-semibold text-foreground">{t("group.home.invite_prompt")}</p>
					<Button type="button" className={authSubmitClass} onClick={goJoin}>
						{t("group.home.join_group_button")}
					</Button>
				</div>
			</WiChatDialogFooter>
		</>
	);

	const createChrome = (
		<>
			<WiChatDialogHeader
				title={t("group.home.customize_title")}
				subtitle={t("group.home.customize_lead")}
				titleId={stepIndex === 1 ? GROUP_ONBOARDING_TITLE_ID : undefined}
			/>
			<WiChatDialogBody>
				<div className="space-y-5 pb-4 sm:space-y-6 sm:pb-5">
					<GroupIconUpload
						value={groupIconFile}
						onChange={setGroupIconFile}
						displayName={watchedName}
					/>
					<Form {...form}>
						<form
							id={GROUP_CREATE_FORM_ID}
							className={authFormClass}
							onSubmit={form.handleSubmit((values) => {
								setCreateBusy(true);
								createGroup.mutate(values, {
									onSuccess: async (group) => {
										try {
											if (groupIconFile) {
												const updated = await uploadGroupIcon(group.id, groupIconFile);
												queryClient.setQueryData<Group[]>([Q_MY_GROUPS], (prev) => {
													const list = prev ?? [];
													const has = list.some((g) => g.id === updated.id);
													if (has) {
														return list.map((g) =>
															g.id === updated.id ? { ...g, logo_url: updated.logo_url } : g,
														);
													}
													return [...list, updated];
												});
											}
											openGroupChannel(group.id, group.default_channel_id);
										} catch (err) {
											try {
												await leaveGroup(group.id);
											} catch {
												// best-effort rollback so create can be retried
											}
											void queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
											form.setError("root", {
												message: translateError(t, err),
											});
										} finally {
											setCreateBusy(false);
										}
									},
									onError: (err) => {
										setCreateBusy(false);
										form.setError("root", { message: translateError(t, err) });
									},
								});
							})}
						>
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
							{form.formState.errors.root?.message ? (
								<p className="text-destructive text-[11px] font-medium">
									{form.formState.errors.root.message}
								</p>
							) : null}
						</form>
					</Form>
				</div>
			</WiChatDialogBody>
			<WiChatDialogFooter>
				<WiChatDialogFooterActions
					backLabel={t("auth.back")}
					onBack={backToPick}
					primaryLabel={
						createBusy || createGroup.isPending
							? t("group.home.creating")
							: t("group.home.create_submit")
					}
					primaryType="submit"
					primaryForm={GROUP_CREATE_FORM_ID}
					primaryDisabled={createBusy || createGroup.isPending}
					backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
					primaryClassName={cn(authSubmitClass, authFormFooterButtonClass)}
				/>
			</WiChatDialogFooter>
		</>
	);

	const joinChrome = (
		<>
			<WiChatDialogHeader
				title={t("group.home.join_title")}
				subtitle={t("group.home.join_lead")}
				titleId={stepIndex === 2 ? GROUP_ONBOARDING_TITLE_ID : undefined}
			/>
			<WiChatDialogBody>
				<div className="space-y-4 pb-4 sm:pb-5">
					{joinable.isPending ? (
						<p className="text-center text-sm text-muted-foreground">{t("group.home.loading")}</p>
					) : (joinable.data ?? []).length === 0 ? (
						<div className="space-y-4 rounded-xl border border-dashed border-border bg-muted/50 px-4 py-8 text-center">
							<p className="text-sm text-muted-foreground">{t("group.home.join_empty")}</p>
							<Button type="button" className={authSubmitClass} onClick={() => startCreate()}>
								{t("group.home.create")}
							</Button>
						</div>
					) : (
						<ul className="divide-y divide-border overflow-hidden rounded-xl border border-border">
							{(joinable.data ?? []).map((g) => (
								<li
									key={g.id}
									className="flex items-center justify-between gap-3 bg-card px-4 py-3.5"
								>
									<span className="font-medium">{g.name}</span>
									<Button
										type="button"
										size="sm"
										disabled={joinGroup.isPending}
										onClick={() =>
											joinGroup.mutate(g.id, {
												onSuccess: async () => {
													const groups = await queryClient.fetchQuery(
														groupQuery.myGroupsQueryOpts(),
													);
													const joined = groups.find((row) => row.id === g.id);
													if (joined?.default_channel_id) {
														openGroupChannel(joined.id, joined.default_channel_id);
														return;
													}
													finish();
												},
											})
										}
									>
										{t("group.home.join_action")}
									</Button>
								</li>
							))}
						</ul>
					)}
				</div>
			</WiChatDialogBody>
			<WiChatDialogFooter>
				<Button type="button" variant="ghost" className="px-2" onClick={backToPick}>
					<ArrowLeft className="mr-1 h-4 w-4" aria-hidden />
					{t("auth.back")}
				</Button>
			</WiChatDialogFooter>
		</>
	);

	const chrome = (
		<WiChatDialogStepCarousel activeIndex={stepIndex}>
			{pickChrome}
			{createChrome}
			{joinChrome}
		</WiChatDialogStepCarousel>
	);

	if (presentation === "dialog") {
		return chrome;
	}

	return (
		<div
			className={cn(
				groupOnboardingCardShellClass,
				"flex h-auto w-full max-h-[min(90vh,44rem)] flex-col overflow-hidden",
			)}
		>
			{chrome}
		</div>
	);
}

type GroupOnboardingDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
};

export function GroupOnboardingDialog({ open, onOpenChange }: GroupOnboardingDialogProps) {
	return (
		<WiChatDialog open={open} onOpenChange={onOpenChange} titleId={GROUP_ONBOARDING_TITLE_ID}>
			<GroupOnboardingFlow presentation="dialog" onComplete={() => onOpenChange(false)} />
		</WiChatDialog>
	);
}
