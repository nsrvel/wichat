"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Lock, Smile } from "lucide-react";
import { useCallback, useId, useRef } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { z } from "zod";
import { ExpressionPickerPopover } from "@/components/expression-picker/expression-picker-popover";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import {
	WiChatDialog,
	WiChatDialogBody,
	WiChatDialogFooter,
	WiChatDialogFooterActions,
	WiChatDialogHeader,
} from "@/components/ui/wichat-dialog";
import {
	authFieldLabelClass,
	authFormClass,
	authInputClass,
	authSubmitClass,
	authSubmitGhostClass,
} from "@/features/auth/components/auth-shell";
import { cn } from "@/lib/utils";

const CREATE_CATEGORY_FORM_ID = "wichat-create-category-form";
const CREATE_CATEGORY_TITLE_ID = "wichat-create-category-title";
const authFormFooterButtonClass = "w-auto min-w-[7.5rem] max-w-[55%] shrink-0 px-8";

const CreateCategorySchema = z.object({
	name: z
		.string()
		.trim()
		.min(1, "channel.create_category_dialog.validation_required")
		.max(64, "channel.create_category_dialog.validation_max"),
	isPrivate: z.boolean(),
});

export type CreateCategoryFormValues = z.infer<typeof CreateCategorySchema>;

export type CreateCategoryTarget = {
	groupId: string;
	groupName?: string;
};

type CreateCategoryDialogProps = {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	target: CreateCategoryTarget | null;
	onSubmitCategory?: (values: CreateCategoryFormValues, target: CreateCategoryTarget) => void;
};

export function CreateCategoryDialog({
	open,
	onOpenChange,
	target,
	onSubmitCategory,
}: CreateCategoryDialogProps) {
	const { t } = useTranslation();
	const privateSwitchId = useId();
	const inputRef = useRef<HTMLInputElement>(null);

	const form = useForm<CreateCategoryFormValues>({
		resolver: zodResolver(CreateCategorySchema),
		defaultValues: {
			name: "",
			isPrivate: false,
		},
	});

	const watchedName = useWatch({ control: form.control, name: "name" }) ?? "";
	const watchedPrivate = useWatch({ control: form.control, name: "isPrivate" }) ?? false;
	const canSubmit = watchedName.trim().length > 0;

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
		<WiChatDialog
			open={open}
			onOpenChange={handleOpenChange}
			titleId={CREATE_CATEGORY_TITLE_ID}
		>
			<WiChatDialogHeader
				title={t("channel.create_category_dialog.title")}
				titleId={CREATE_CATEGORY_TITLE_ID}
			/>

			<WiChatDialogBody>
				<Form {...form}>
					<form
						id={CREATE_CATEGORY_FORM_ID}
						className={cn(authFormClass, "pb-2")}
						onSubmit={form.handleSubmit((values) => {
							if (target && onSubmitCategory) {
								onSubmitCategory(values, target);
							} else {
								toast.info(
									t("channel.create_category_dialog.preview_toast", {
										name: values.name,
									}),
								);
							}
							handleOpenChange(false);
						})}
					>
						<FormField
							control={form.control}
							name="name"
							render={({ field }) => (
								<FormItem className="space-y-1.5">
									<FormLabel className={authFieldLabelClass}>
										{t("channel.create_category_dialog.name_label")}
										<span className="text-destructive" aria-hidden>
											{" *"}
										</span>
									</FormLabel>
									<FormControl>
										<div className="relative flex items-center">
											<Input
												{...field}
												ref={(node) => {
													field.ref(node);
													inputRef.current = node;
												}}
												maxLength={64}
												autoFocus
												placeholder={t(
													"channel.create_category_dialog.name_placeholder",
												)}
												className={cn(authInputClass, "pr-11")}
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
														aria-label={t(
															"channel.create_category_dialog.emoji_picker",
														)}
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
										<Lock
											className="size-4 shrink-0 text-muted-foreground"
											aria-hidden
										/>
										<span>
											{t("channel.create_category_dialog.private_title")}
										</span>
									</label>
									<p className="text-xs leading-relaxed text-muted-foreground">
										{t("channel.create_category_dialog.private_description")}
									</p>
								</div>
								<Switch
									id={privateSwitchId}
									checked={watchedPrivate}
									onCheckedChange={(checked) =>
										form.setValue("isPrivate", checked, { shouldDirty: true })
									}
									aria-label={t("channel.create_category_dialog.private_title")}
									className="mt-0.5"
								/>
							</div>
						</div>
					</form>
				</Form>
			</WiChatDialogBody>

			<WiChatDialogFooter>
				<WiChatDialogFooterActions
					backLabel={t("channel.create_category_dialog.cancel")}
					onBack={() => handleOpenChange(false)}
					primaryLabel={t("channel.create_category_dialog.submit")}
					primaryType="submit"
					primaryForm={CREATE_CATEGORY_FORM_ID}
					primaryDisabled={!canSubmit}
					backClassName={cn(authSubmitGhostClass, authFormFooterButtonClass)}
					primaryClassName={cn(authSubmitClass, authFormFooterButtonClass)}
				/>
			</WiChatDialogFooter>
		</WiChatDialog>
	);
}

