"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ChevronRight } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { TextField } from "@/components/form/text-field";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
import { groupNameSchema, type Group } from "@wichat/contracts";
import { cn } from "@/lib/utils";
import { normalizeInviteToken } from "@/features/invite/lib/normalize-invite-token";
import { channelPath } from "@/lib/app-routes";
import { Q_MY_GROUPS } from "@/services/group/group.key";
import { uploadGroupIcon } from "@/services/group/group.service";
import { useCreateGroupMutation } from "@/services/group/group.mutation";
import { redeemInvite } from "@/services/invite/invite.service";

const GROUP_ONBOARDING_TITLE_ID = "group-onboarding-title";
const GROUP_CREATE_FORM_ID = "wichat-group-create-form";
const authFormFooterButtonClass = "w-auto min-w-[7.5rem] max-w-[55%] shrink-0 px-8";

const createSchema = z.object({ name: groupNameSchema });

export type GroupOnboardingMode = "pick" | "create" | "join";

type Mode = GroupOnboardingMode;

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
	"flex size-10 shrink-0 items-center justify-center text-[1.625rem] leading-none";

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
	initialMode?: Mode;
};

export function GroupOnboardingFlow({
	onComplete,
	presentation = "page",
	initialMode = "pick",
}: GroupOnboardingFlowProps) {
	const { t } = useTranslation();
	const [mode, setMode] = useState<Mode>(initialMode);
	const stepIndex = modeStepIndex(mode);
	const router = useRouter();
	const queryClient = useQueryClient();
	const createGroup = useCreateGroupMutation();
	const [inviteCode, setInviteCode] = useState("");
	const joinByInvite = useMutation({
		mutationFn: (token: string) => redeemInvite(token),
		onSuccess: async (result) => {
			await queryClient.invalidateQueries({ queryKey: [Q_MY_GROUPS] });
			if (result.channel_id) {
				openGroupChannel(result.group_id, result.channel_id);
				return;
			}
			finish();
		},
	});
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
						<ChevronRight className="size-5 shrink-0 text-muted-foreground" aria-hidden />
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
										<ChevronRight className="size-5 shrink-0 text-muted-foreground" aria-hidden />
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
											const owned: Group = {
												...group,
												role: group.role ?? "owner",
											};
											queryClient.setQueryData<Group[]>([Q_MY_GROUPS], (prev) => [
												...(prev ?? []).filter((g) => g.id !== owned.id),
												owned,
											]);

											if (groupIconFile) {
												try {
													const updated = await uploadGroupIcon(group.id, groupIconFile);
													queryClient.setQueryData<Group[]>([Q_MY_GROUPS], (prev) => {
														const list = prev ?? [];
														return list.map((g) =>
															g.id === updated.id
																? {
																		...g,
																		logo_url: updated.logo_url,
																		role: updated.role ?? g.role ?? "owner",
																	}
																: g,
														);
													});
												} catch {
													// Group already created — continue without logo (e.g. MinIO off locally).
												}
											}
											openGroupChannel(group.id, group.default_channel_id);
										} catch (err) {
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
					<div className="space-y-2">
						<label htmlFor="group-invite-code" className={authFieldLabelClass}>
							{t("group.home.join_code_label")}
						</label>
						<Input
							id="group-invite-code"
							value={inviteCode}
							onChange={(e) => setInviteCode(e.target.value)}
							placeholder={t("group.home.join_code_placeholder")}
							className={authInputClass}
							autoComplete="off"
							disabled={joinByInvite.isPending}
						/>
					</div>
					{joinByInvite.isError ? (
						<p className="text-destructive text-[11px] font-medium">
							{translateError(t, joinByInvite.error)}
						</p>
					) : null}
				</div>
			</WiChatDialogBody>
			<WiChatDialogFooter>
				<WiChatDialogFooterActions
					backLabel={t("auth.back")}
					onBack={backToPick}
					primaryLabel={
						joinByInvite.isPending ? t("group.home.joining") : t("group.home.join_code_submit")
					}
					primaryType="button"
					primaryDisabled={joinByInvite.isPending || !inviteCode.trim()}
					onPrimary={() => {
						const token = normalizeInviteToken(inviteCode);
						if (!token) {
							return;
						}
						joinByInvite.mutate(token);
					}}
					backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
					primaryClassName={cn(authSubmitClass, authFormFooterButtonClass)}
				/>
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
				"flex h-auto w-full max-h-[90vh] flex-col overflow-hidden",
			)}
		>
			{chrome}
		</div>
	);
}

type GroupOnboardingDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	entryMode?: GroupOnboardingMode;
};

export function GroupOnboardingDialog({
	open,
	onOpenChange,
	entryMode = "pick",
}: GroupOnboardingDialogProps) {
	return (
		<WiChatDialog open={open} onOpenChange={onOpenChange} titleId={GROUP_ONBOARDING_TITLE_ID}>
			{open ? (
				<GroupOnboardingFlow
					key={entryMode}
					initialMode={entryMode}
					presentation="dialog"
					onComplete={() => onOpenChange(false)}
				/>
			) : null}
		</WiChatDialog>
	);
}
