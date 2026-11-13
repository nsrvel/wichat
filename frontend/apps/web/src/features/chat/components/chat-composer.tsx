"use client";

import { AtSign, Paperclip, Plus, Send, Smile, Type } from "lucide-react";
import { useCallback, useId, useRef, useState, type KeyboardEvent } from "react";
import { useTranslation } from "react-i18next";
import { cn } from "@/lib/utils";

const COMPOSER_ACTIONS = [
	{ icon: Type, labelKey: "chat.formatting" as const },
	{ icon: Smile, labelKey: "chat.emoji" as const },
	{ icon: Paperclip, labelKey: "chat.attach" as const },
	{ icon: AtSign, labelKey: "chat.mention" as const },
	{ icon: Plus, labelKey: "chat.more" as const },
] as const;

type ChatComposerProps = {
	className?: string;
	placeholder?: string;
	value?: string;
	onChange?: (value: string) => void;
	onSubmit?: (value: string) => void;
	disabled?: boolean;
	submitDisabled?: boolean;
};

export function ChatComposer({
	className,
	placeholder,
	value,
	onChange,
	onSubmit,
	disabled = false,
	submitDisabled,
}: ChatComposerProps) {
	const { t } = useTranslation();
	const inputId = useId();
	const textareaRef = useRef<HTMLTextAreaElement>(null);
	const [internalDraft, setInternalDraft] = useState("");
	const controlled = value !== undefined;
	const draft = controlled ? (value ?? "") : internalDraft;
	const resolvedPlaceholder = placeholder ?? t("chat.composer_placeholder");
	const canSend = draft.trim().length > 0 && !disabled && submitDisabled !== true;

	const resizeTextarea = useCallback(() => {
		const el = textareaRef.current;
		if (!el) {
			return;
		}
		el.style.height = "auto";
		el.style.height = `${Math.min(el.scrollHeight, 160)}px`;
	}, []);

	const submit = useCallback(() => {
		const text = draft.trim();
		if (!text || !canSend) {
			return;
		}
		onSubmit?.(text);
		if (controlled) {
			onChange?.("");
		} else {
			setInternalDraft("");
			if (textareaRef.current) {
				textareaRef.current.value = "";
			}
			resizeTextarea();
		}
	}, [canSend, controlled, draft, onChange, onSubmit, resizeTextarea]);

	const handleKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>) => {
		if (event.key === "Enter" && !event.shiftKey) {
			event.preventDefault();
			submit();
		}
	};

	return (
		<div
			className={cn(
				"mx-auto w-full max-w-4xl shrink-0 px-4 pb-4 pt-1 sm:px-5 lg:max-w-5xl lg:px-6 lg:pb-5",
				className,
			)}
		>
			<div className="flex min-h-11 items-end gap-2 rounded-xl border border-chat-bubble-border bg-chat-bubble px-3 py-1.5">
				<label htmlFor={inputId} className="sr-only">
					{resolvedPlaceholder}
				</label>
				<textarea
					ref={textareaRef}
					id={inputId}
					rows={1}
					disabled={disabled}
					value={controlled ? value : undefined}
					defaultValue={controlled ? undefined : ""}
					onChange={(event) => {
						const next = event.target.value;
						if (controlled) {
							onChange?.(next);
						} else {
							setInternalDraft(next);
						}
						resizeTextarea();
					}}
					onKeyDown={handleKeyDown}
					placeholder={resolvedPlaceholder}
					className="thin-scroll max-h-40 min-h-6 min-w-0 flex-1 resize-none overflow-y-auto bg-transparent py-1 text-[15px] leading-6 text-foreground placeholder:text-muted-foreground focus:outline-none disabled:opacity-50"
				/>
				<div className="flex shrink-0 items-center">
					{COMPOSER_ACTIONS.map(({ icon: Icon, labelKey }) => (
						<button
							key={labelKey}
							type="button"
							disabled={disabled}
							className="flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground disabled:opacity-50"
							aria-label={t(labelKey)}
						>
							<Icon className="size-[18px]" strokeWidth={1.75} />
						</button>
					))}
					<span className="mx-1.5 h-6 w-px shrink-0 bg-chat-bubble-border" aria-hidden />
					<button
						type="button"
						disabled={!canSend}
						onClick={submit}
						className="flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground disabled:opacity-40"
						aria-label={t("chat.send")}
						title={submitDisabled ? t("chat.send_soon") : undefined}
					>
						<Send className="size-[18px]" strokeWidth={1.75} />
					</button>
				</div>
			</div>
		</div>
	);
}
