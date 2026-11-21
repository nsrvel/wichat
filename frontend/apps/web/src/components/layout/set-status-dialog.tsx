"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Smile, X } from "lucide-react";
import { useId, useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { z } from "zod";
import { ExpressionPickerPopover } from "@/components/expression-picker/expression-picker-popover";
import { Button } from "@/components/ui/button";
import { CustomSelect, type SelectOption } from "@/components/ui/custom-select";
import { Form, FormControl, FormField, FormItem, FormMessage } from "@/components/ui/form";
import { UserAvatar } from "@/components/ui/user-avatar";
import {
	WiChatDialog,
	WiChatDialogBody,
	WiChatDialogFooter,
	WiChatDialogHeader,
} from "@/components/ui/wichat-dialog";
import { translateError } from "@/integrations/i18n/translate-error";
import {
	computeStatusNoteExpiry,
	formatStatusNoteLabel,
	type StatusClearDuration,
	type StatusNote,
	type UserPresenceStatus,
} from "@/lib/user-status";

const statusNoteFormSchema = z.object({
	text: z.string().trim().max(80),
	emoji: z.string().trim().max(16),
	duration: z.enum(["today", "4h", "1h", "30m", "never"]),
});

type StatusNoteFormValues = z.infer<typeof statusNoteFormSchema>;

type SetStatusDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	displayName: string;
	username: string;
	avatarUrl: string;
	bannerColor?: string;
	initials: string;
	status: UserPresenceStatus;
	/** Effective presence dot on the preview avatar (WS); falls back to manual `status`. */
	indicatorStatus?: UserPresenceStatus | "offline";
	initialNote: StatusNote;
	onSave: (note: StatusNote) => Promise<void>;
};

function SetStatusDialogBody({
	onClose,
	displayName,
	username,
	avatarUrl,
	bannerColor,
	initials,
	status,
	indicatorStatus,
	initialNote,
	onSave,
	titleId,
}: Omit<SetStatusDialogProps, "open" | "onOpenChange"> & {
	onClose: () => void;
	titleId: string;
}) {
	const { t } = useTranslation();
	const [emojiPickerOpen, setEmojiPickerOpen] = useState(false);

	const form = useForm<StatusNoteFormValues>({
		resolver: zodResolver(statusNoteFormSchema),
		defaultValues: {
			text: initialNote.text,
			emoji: initialNote.emoji,
			duration: "today",
		},
	});

	const textDraft = useWatch({ control: form.control, name: "text" }) ?? "";
	const emojiDraft = useWatch({ control: form.control, name: "emoji" }) ?? "";
	const duration = useWatch({ control: form.control, name: "duration" }) ?? "today";

	const previewLabel = formatStatusNoteLabel({
		text: textDraft.trim(),
		emoji: emojiDraft,
		expiresAt: null,
	});

	const durationOptions: SelectOption[] = [
		{ value: "today", label: t("shell.user_dock.clear_after_today") },
		{ value: "4h", label: t("shell.user_dock.clear_after_4h") },
		{ value: "1h", label: t("shell.user_dock.clear_after_1h") },
		{ value: "30m", label: t("shell.user_dock.clear_after_30m") },
		{ value: "never", label: t("shell.user_dock.clear_after_never") },
	];

	const handleSubmit = form.handleSubmit(async (values) => {
		const text = values.text.trim();
		const emoji = values.emoji.trim();
		try {
			await onSave({
				text,
				emoji,
				expiresAt: text || emoji ? computeStatusNoteExpiry(values.duration) : null,
			});
			onClose();
		} catch (err) {
			toast.error(translateError(t, err));
		}
	});

	const formId = `${titleId}-form`;
	const isPending = form.formState.isSubmitting;

	return (
		<>
			<WiChatDialogHeader
				title={t("shell.user_dock.set_status_title")}
				titleId={titleId}
				className="text-left"
			/>

			<WiChatDialogBody>
				<Form {...form}>
					<form id={formId} onSubmit={handleSubmit} className="flex flex-col gap-5">
						{/* Live profile preview card */}
						<div className="mx-auto w-full max-w-[300px] overflow-visible rounded-2xl border border-border/70 bg-card text-card-foreground shadow-sm">
							<div
								className="h-[72px] overflow-hidden rounded-t-2xl bg-popover"
								style={bannerColor ? { backgroundColor: bannerColor } : undefined}
							/>
							<div className="px-5 pb-5">
								<div className="-mt-[36px] flex items-center gap-2.5">
									<UserAvatar
										src={avatarUrl}
										initials={initials}
										size="xl"
										status={indicatorStatus ?? status}
										ringClassName="ring-[5px] ring-card"
									/>

									<div className="min-w-0 flex-1">
										<div className="w-fit max-w-full rounded-2xl bg-muted px-3.5 py-2.5 text-xs shadow-[0_6px_16px_rgba(0,0,0,0.28)]">
											<p className="line-clamp-2 break-words leading-snug text-foreground/90">
												{previewLabel || (
													<span className="text-muted-foreground">
														{t("shell.user_dock.status_note_placeholder")}
													</span>
												)}
											</p>
										</div>
									</div>
								</div>

								<div className="mt-3.5 min-w-0">
									<p className="truncate text-base font-bold leading-tight tracking-tight text-foreground">
										{displayName}
									</p>
									{username ? (
										<p className="mt-0.5 truncate text-xs font-medium text-muted-foreground">
											@{username}
										</p>
									) : null}
								</div>
							</div>
						</div>

						{/* Status input */}
						<FormField
							control={form.control}
							name="text"
							render={({ field }) => (
								<FormItem className="space-y-2 pb-1">
									<label
										htmlFor={`${titleId}-input`}
										className="block text-[13px] font-medium text-foreground"
									>
										{t("shell.user_dock.set_status_label")}
									</label>
									<FormControl>
										<div className="flex items-center gap-2 rounded-xl border border-input bg-field px-2.5 py-1.5 focus-within:border-ring focus-within:ring-2 focus-within:ring-ring/30">
											<ExpressionPickerPopover
												only="emoji"
												open={emojiPickerOpen}
												onOpenChange={setEmojiPickerOpen}
												closeOnSelect
												side="top"
												align="start"
												onSelect={(selection) => {
													if (selection.kind === "emoji") {
														form.setValue("emoji", selection.emoji, {
															shouldDirty: true,
															shouldValidate: true,
														});
													}
												}}
												trigger={
													<button
														type="button"
														aria-label={t("shell.user_dock.pick_emoji")}
														disabled={isPending}
														className="flex size-8 shrink-0 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-foreground/10 hover:text-foreground disabled:opacity-50"
													>
														{emojiDraft ? (
															<span className="text-lg leading-none">{emojiDraft}</span>
														) : (
															<Smile className="size-4" aria-hidden />
														)}
													</button>
												}
											/>
											<input
												{...field}
												id={`${titleId}-input`}
												type="text"
												maxLength={80}
												autoFocus
												disabled={isPending}
												placeholder={t("shell.user_dock.status_note_placeholder")}
												className="min-w-0 flex-1 bg-transparent text-sm text-foreground placeholder:text-placeholder-foreground focus:outline-none disabled:opacity-50"
											/>
											{textDraft || emojiDraft ? (
												<button
													type="button"
													aria-label={t("shell.user_dock.status_note_clear")}
													disabled={isPending}
													onClick={() => {
														form.setValue("text", "", { shouldDirty: true });
														form.setValue("emoji", "", { shouldDirty: true });
													}}
													className="flex size-7 shrink-0 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-foreground/10 hover:text-foreground disabled:opacity-50"
												>
													<X className="size-4" aria-hidden />
												</button>
											) : null}
										</div>
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>
					</form>
				</Form>
			</WiChatDialogBody>

			<WiChatDialogFooter className="flex items-center justify-between gap-3">
				<CustomSelect
					options={durationOptions}
					value={duration}
					placeholder={t("shell.user_dock.clear_after_today")}
					onChange={(val) =>
						form.setValue("duration", val as StatusClearDuration, { shouldDirty: true })
					}
					portaled
					compactTrigger
					triggerClassName="h-10 rounded-xl px-3.5 text-sm font-medium text-muted-foreground hover:bg-foreground/5 hover:text-foreground"
					menuClassName="z-[90] min-w-[180px]"
				/>
				<Button
					type="submit"
					form={formId}
					disabled={isPending}
					className="h-10 min-w-[88px] rounded-xl px-5 font-medium"
				>
					{t("shell.user_dock.set_status_save")}
				</Button>
			</WiChatDialogFooter>
		</>
	);
}

export function SetStatusDialog({ open, onOpenChange, ...props }: SetStatusDialogProps) {
	const titleId = useId();

	return (
		<WiChatDialog
			elevated
			open={open}
			onOpenChange={onOpenChange}
			titleId={titleId}
			className="max-w-[440px]"
		>
			<SetStatusDialogBody {...props} titleId={titleId} onClose={() => onOpenChange(false)} />
		</WiChatDialog>
	);
}
