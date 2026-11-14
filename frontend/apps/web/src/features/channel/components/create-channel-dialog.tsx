"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Hash, Lock, Megaphone, MessagesSquare, Smile, Volume2 } from "lucide-react";
import { useCallback, useEffect, useId, useRef } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { z } from "zod";
import { ExpressionPickerPopover } from "@/components/expression-picker/expression-picker-popover";
import {
	Form,
	FormControl,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import {
	WiChatDialog,
	WiChatDialogBody,
	WiChatDialogFooter,
	WiChatDialogFooterActions,
	WiChatDialogHeader,
} from "@/components/ui/wichat-dialog";
import { authFieldLabelClass, authFormClass, authInputClass, authSubmitClass, authSubmitGhostClass } from "@/lib/form-field-styles";
import { cn } from "@/lib/utils";

const CREATE_CHANNEL_FORM_ID = "wichat-create-channel-form";
const CREATE_CHANNEL_TITLE_ID = "wichat-create-channel-title";
const authFormFooterButtonClass = "w-auto min-w-[7.5rem] max-w-[55%] shrink-0 px-8";

const CHANNEL_TYPE_OPTIONS = [
	{ value: "text_only", icon: Hash },
	{ value: "voice_only", icon: Volume2 },
	{ value: "forum", icon: MessagesSquare },
	{ value: "broadcast", icon: Megaphone },
] as const;

export type CreatableChannelType = (typeof CHANNEL_TYPE_OPTIONS)[number]["value"];

const CreateChannelSchema = z.object({
	type: z.enum(["text_only", "voice_only", "forum", "broadcast"]),
	name: z
		.string()
		.trim()
		.min(1, "channel.create_channel_dialog.validation_required")
		.max(64, "channel.create_channel_dialog.validation_max"),
	categoryId: z.string().nullable(),
	isPrivate: z.boolean(),
});

export type CreateChannelFormValues = z.infer<typeof CreateChannelSchema>;

export type CreateChannelTarget = {
	groupId: string;
	groupName?: string;
	defaultCategoryId?: string | null;
};

type CreateChannelDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	target: CreateChannelTarget | null;
	onSubmitChannel?: (values: CreateChannelFormValues, target: CreateChannelTarget) => void;
};

export function CreateChannelDialog({
	open,
	onOpenChange,
	target,
	onSubmitChannel,
}: CreateChannelDialogProps) {
	const { t } = useTranslation();
	const privateSwitchId = useId();
	const inputRef = useRef<HTMLInputElement>(null);

	const form = useForm<CreateChannelFormValues>({
		resolver: zodResolver(CreateChannelSchema),
		defaultValues: {
			type: "text_only",
			name: "",
			categoryId: target?.defaultCategoryId ?? null,
			isPrivate: false,
		},
	});

	useEffect(() => {
		if (open) {
			form.reset({
				type: "text_only",
				name: "",
				categoryId: target?.defaultCategoryId ?? null,
				isPrivate: false,
			});
		}
	}, [form, open, target?.defaultCategoryId]);

	const watchedType = useWatch({ control: form.control, name: "type" }) ?? "text_only";
	const watchedName = useWatch({ control: form.control, name: "name" }) ?? "";
	const watchedPrivate = useWatch({ control: form.control, name: "isPrivate" }) ?? false;
	const canSubmit = watchedName.trim().length > 0;

	const SelectedTypeIcon =
		CHANNEL_TYPE_OPTIONS.find((opt) => opt.value === watchedType)?.icon ?? Hash;

	const handleOpenChange = useCallback(
		(next: boolean) => {
			if (!next) {
				form.reset();
			}
			onOpenChange(next);
		},
		[form, onOpenChange],
	);

	const insertEmoji = useCallback(
		(emoji: string) => {
			const el = inputRef.current;
			const current = form.getValues("name") ?? "";
			const start = el?.selectionStart ?? current.length;
			const end = el?.selectionEnd ?? current.length;
			const next = `${current.slice(0, start)}${emoji}${current.slice(end)}`;
			form.setValue("name", next, { shouldDirty: true, shouldValidate: true });
			requestAnimationFrame(() => {
				const targetEl = inputRef.current;
				if (!targetEl) {
					return;
				}
				const caret = start + emoji.length;
				targetEl.focus();
				targetEl.setSelectionRange(caret, caret);
			});
		},
		[form],
	);

	return (
		<WiChatDialog open={open} onOpenChange={handleOpenChange} titleId={CREATE_CHANNEL_TITLE_ID}>
			<WiChatDialogHeader
				title={t("channel.create_channel_dialog.title")}
				titleId={CREATE_CHANNEL_TITLE_ID}
			/>

			<WiChatDialogBody>
				<Form {...form}>
					<form
						id={CREATE_CHANNEL_FORM_ID}
						className={cn(authFormClass, "pb-2")}
						onSubmit={form.handleSubmit((values) => {
							if (target && onSubmitChannel) {
								onSubmitChannel(values, target);
							} else {
								toast.info(
									t("channel.create_channel_dialog.preview_toast", {
										name: values.name,
										type: t(`channel.create_channel_dialog.types.${values.type}.label`),
									}),
								);
							}
							handleOpenChange(false);
						})}
					>
						<div className="space-y-1.5">
							<span className={authFieldLabelClass}>
								{t("channel.create_channel_dialog.type_label")}
							</span>
							<div
								role="radiogroup"
								aria-label={t("channel.create_channel_dialog.type_label")}
								className="space-y-1"
							>
								{CHANNEL_TYPE_OPTIONS.map(({ value, icon: Icon }) => {
									const selected = watchedType === value;
									return (
										<button
											key={value}
											type="button"
											role="radio"
											aria-checked={selected}
											onClick={() => form.setValue("type", value, { shouldDirty: true })}
											className={cn(
												"flex w-full items-center gap-3 rounded-lg px-2 py-2 text-left transition-colors hover:bg-muted/40",
												selected ? "bg-muted/35" : "",
											)}
										>
											<span
												className={cn(
													"flex size-[18px] shrink-0 items-center justify-center rounded-full border-2 transition-colors",
													selected ? "border-primary bg-primary" : "border-muted-foreground/50",
												)}
												aria-hidden
											>
												{selected ? (
													<span className="size-1.5 rounded-full bg-primary-foreground" />
												) : null}
											</span>
											<div className="min-w-0 flex-1 space-y-0.5">
												<div className="flex items-center gap-1.5">
													<Icon
														className={cn(
															"size-[18px] shrink-0",
															selected ? "text-foreground" : "text-muted-foreground",
														)}
														strokeWidth={2}
														aria-hidden
													/>
													<p className="text-sm font-semibold text-foreground">
														{t(`channel.create_channel_dialog.types.${value}.label`)}
													</p>
												</div>
												<p className="text-xs leading-snug text-muted-foreground">
													{t(`channel.create_channel_dialog.types.${value}.description`)}
												</p>
											</div>
										</button>
									);
								})}
							</div>
						</div>

						<FormField
							control={form.control}
							name="name"
							render={({ field }) => (
								<FormItem className="space-y-1.5">
									<FormLabel className={authFieldLabelClass}>
										{t("channel.create_channel_dialog.name_label")}
										<span className="text-destructive" aria-hidden>
											{" *"}
										</span>
									</FormLabel>
									<FormControl>
										<div className="relative flex items-center">
											<SelectedTypeIcon
												className="pointer-events-none absolute left-3.5 size-4 text-muted-foreground"
												strokeWidth={2}
												aria-hidden
											/>
											<Input
												{...field}
												ref={(node) => {
													field.ref(node);
													inputRef.current = node;
												}}
												maxLength={64}
												autoFocus
												placeholder={t("channel.create_channel_dialog.name_placeholder")}
												className={cn(authInputClass, "pl-10 pr-11 sm:pl-10 sm:pr-11")}
											/>
											<ExpressionPickerPopover
												only="emoji"
												side="bottom"
												align="end"
												sideOffset={6}
												onSelect={(selection) => {
													if (selection.kind === "emoji") {
														insertEmoji(selection.emoji);
													}
												}}
												trigger={
													<button
														type="button"
														aria-label={t("channel.create_channel_dialog.emoji_picker")}
														className="absolute right-2 flex size-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground data-[popup-open]:bg-accent data-[popup-open]:text-foreground"
													>
														<Smile className="size-[18px]" strokeWidth={1.8} />
													</button>
												}
											/>
										</div>
									</FormControl>
									<FormMessage />
								</FormItem>
							)}
						/>

						<div className="rounded-xl border border-border/70 bg-muted/30 p-3.5">
							<div className="flex items-start justify-between gap-4">
								<div className="space-y-1">
									<label
										htmlFor={privateSwitchId}
										className="flex cursor-pointer items-center gap-1.5 text-sm font-semibold text-foreground"
									>
										<Lock className="size-4 shrink-0 text-muted-foreground" aria-hidden />
										<span>{t("channel.create_channel_dialog.private_title")}</span>
									</label>
									<p className="text-xs leading-relaxed text-muted-foreground">
										{t("channel.create_channel_dialog.private_description")}
									</p>
								</div>
								<Switch
									id={privateSwitchId}
									checked={watchedPrivate}
									onCheckedChange={(checked) =>
										form.setValue("isPrivate", checked, { shouldDirty: true })
									}
									aria-label={t("channel.create_channel_dialog.private_title")}
									className="mt-0.5"
								/>
							</div>
						</div>
					</form>
				</Form>
			</WiChatDialogBody>

			<WiChatDialogFooter>
				<WiChatDialogFooterActions
					backLabel={t("channel.create_channel_dialog.cancel")}
					onBack={() => handleOpenChange(false)}
					primaryLabel={t("channel.create_channel_dialog.submit")}
					primaryType="submit"
					primaryForm={CREATE_CHANNEL_FORM_ID}
					primaryDisabled={!canSubmit}
					backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
					primaryClassName={cn(authSubmitClass, authFormFooterButtonClass)}
				/>
			</WiChatDialogFooter>
		</WiChatDialog>
	);
}
