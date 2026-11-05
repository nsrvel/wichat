"use client";

import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { FieldLabelContent } from "@/components/form/field-label";
import { authOtpCellClass } from "@/features/auth/components/auth-shell";
import { cn } from "@/lib/utils";
import { useCallback, useRef, useState } from "react";
import type { FieldPath, FieldValues } from "react-hook-form";
import { useTranslation } from "react-i18next";

type OtpFieldProps<T extends FieldValues> = {
	control: import("react-hook-form").Control<T>;
	name: FieldPath<T>;
	label: string;
	length?: number;
	labelClassName?: string;
	required?: boolean;
	hideLabel?: boolean;
};

function normalizeOtpChar(value: string): string {
	return value.replace(/[^a-zA-Z0-9]/g, "").toUpperCase();
}

function normalizeOtpValue(value: string, maxLen: number): string {
	return normalizeOtpChar(value).slice(0, maxLen);
}

function otpCellClass({
	hasValue,
	isActive,
	invalid,
}: {
	hasValue: boolean;
	isActive: boolean;
	invalid: boolean;
}) {
	return cn(
		authOtpCellClass,
		invalid && "!border-destructive/50",
		!invalid && isActive && "!border-foreground/20",
		!invalid && !isActive && hasValue && "!border-foreground/12",
	);
}

/** Visual grouping only (e.g. ABC-123); stored value stays alphanumeric without dash. */
function otpVisualSplitAt(length: number): number | null {
	return length === 6 ? 3 : null;
}

export function OtpField<T extends FieldValues>({
	control,
	name,
	label,
	length = 6,
	labelClassName,
	required = false,
	hideLabel = false,
}: OtpFieldProps<T>) {
	const { t } = useTranslation();
	const inputRefs = useRef<(HTMLInputElement | null)[]>([]);
	const [activeIndex, setActiveIndex] = useState(-1);

	const setRef = useCallback((index: number, el: HTMLInputElement | null) => {
		inputRefs.current[index] = el;
	}, []);

	const focusAt = useCallback(
		(index: number) => {
			const clamped = Math.max(0, Math.min(length - 1, index));
			inputRefs.current[clamped]?.focus();
			setActiveIndex(clamped);
		},
		[length],
	);

	return (
		<FormField
			control={control}
			name={name}
			render={({ field, fieldState }) => {
				const value = normalizeOtpValue(String(field.value ?? ""), length);
				const chars = Array.from({ length }, (_, i) => value[i] ?? "");
				const invalid = !!fieldState.error;

				const commit = (next: string) => {
					field.onChange(normalizeOtpValue(next, length));
				};

				const splitAt = otpVisualSplitAt(length);
				const cellGap = "gap-1.5 sm:gap-2 max-h-[880px]:gap-1";
				const groupGrid = `grid flex-1 grid-cols-3 items-center ${cellGap}`;

				const renderCell = (char: string, index: number) => (
					<input
						key={index}
						ref={(el) => setRef(index, el)}
						type="text"
						inputMode="text"
						autoCapitalize="characters"
						spellCheck={false}
						autoComplete={index === 0 ? "one-time-code" : "off"}
						aria-label={t("forgot_password.otp_cell", {
							current: index + 1,
							total: length,
						})}
						aria-required={required}
						aria-invalid={invalid}
						className={otpCellClass({
							hasValue: Boolean(char),
							isActive: activeIndex === index,
							invalid,
						})}
						value={char}
						maxLength={1}
						onChange={(event) => {
							const segment = normalizeOtpChar(event.target.value);
							const nextChar = segment.slice(-1);
							if (!nextChar) {
								const next = value.slice(0, index) + value.slice(index + 1);
								commit(next);
								return;
							}
							const next = value.slice(0, index) + nextChar + value.slice(index + 1);
							commit(next.slice(0, length));
							if (index < length - 1) {
								focusAt(index + 1);
							}
						}}
						onKeyDown={(event) => {
							if (event.key === "Backspace") {
								if (char) {
									const next = value.slice(0, index) + value.slice(index + 1);
									commit(next);
								} else if (index > 0) {
									const next = value.slice(0, index - 1) + value.slice(index);
									commit(next);
									focusAt(index - 1);
								}
								event.preventDefault();
								return;
							}
							if (event.key === "ArrowLeft") {
								focusAt(index - 1);
								event.preventDefault();
								return;
							}
							if (event.key === "ArrowRight") {
								focusAt(index + 1);
								event.preventDefault();
							}
						}}
						onPaste={(event) => {
							event.preventDefault();
							const pasted = normalizeOtpValue(event.clipboardData.getData("text"), length);
							if (!pasted) {
								return;
							}
							const next = (value.slice(0, index) + pasted).slice(0, length);
							commit(next);
							focusAt(Math.min(index + pasted.length, length - 1));
						}}
						onFocus={(event) => {
							setActiveIndex(index);
							event.target.select();
						}}
						onBlur={() => setActiveIndex(-1)}
					/>
				);

				return (
					<FormItem>
						<FormLabel className={cn(labelClassName, hideLabel && "sr-only")}>
							<FieldLabelContent required={required}>{label}</FieldLabelContent>
						</FormLabel>
						<FormControl>
							<div
								className={cn(
									"flex w-full items-center",
									splitAt !== null ? cellGap : `grid grid-cols-6 ${cellGap}`,
								)}
								role="group"
								aria-label={label}
							>
								{splitAt !== null ? (
									<>
										<div className={groupGrid}>
											{chars.slice(0, splitAt).map((char, i) => renderCell(char, i))}
										</div>
										<span
											className="shrink-0 select-none pb-0.5 text-lg font-medium text-muted-foreground"
											aria-hidden
										>
											-
										</span>
										<div className={groupGrid}>
											{chars.slice(splitAt).map((char, i) => renderCell(char, i + splitAt))}
										</div>
									</>
								) : (
									chars.map((char, index) => renderCell(char, index))
								)}
							</div>
						</FormControl>
						<FormMessage />
					</FormItem>
				);
			}}
		/>
	);
}
