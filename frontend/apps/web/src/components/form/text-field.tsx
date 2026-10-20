"use client";

import {
	FormControl,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { FieldLabelContent } from "@/components/form/field-label";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import type { FieldPath, FieldValues } from "react-hook-form";

type TextFieldProps<T extends FieldValues> = {
	control: import("react-hook-form").Control<T>;
	name: FieldPath<T>;
	label: string;
	placeholder: string;
	autoComplete?: string;
	type?: string;
	inputMode?: React.HTMLAttributes<HTMLInputElement>["inputMode"];
	maxLength?: number;
	/** Strip non-digits on change (e.g. OTP). */
	numericOnly?: boolean;
	hideLabel?: boolean;
	/** Defaults to sentence-case; pass e.g. uppercase tracking for exceptions. */
	labelClassName?: string;
	inputClassName?: string;
	required?: boolean;
};

export function TextField<T extends FieldValues>({
	control,
	name,
	label,
	placeholder,
	autoComplete,
	type = "text",
	inputMode,
	maxLength,
	numericOnly = false,
	hideLabel = false,
	labelClassName,
	inputClassName,
	required = false,
}: TextFieldProps<T>) {
	return (
		<FormField
			control={control}
			name={name}
			render={({ field, fieldState }) => (
				<FormItem>
					<FormLabel className={cn(labelClassName, hideLabel && "sr-only")}>
						<FieldLabelContent required={required}>{label}</FieldLabelContent>
					</FormLabel>
					<FormControl>
						<Input
							type={type}
							inputMode={inputMode}
							maxLength={maxLength}
							autoComplete={autoComplete}
							placeholder={placeholder}
							aria-required={required}
							aria-invalid={!!fieldState.error}
							className={inputClassName}
							{...field}
							onChange={(event) => {
								let next = event.target.value;
								if (numericOnly) {
									next = next.replace(/\D/g, "");
									if (maxLength !== undefined) {
										next = next.slice(0, maxLength);
									}
								}
								field.onChange(next);
							}}
						/>
					</FormControl>
					<FormMessage />
				</FormItem>
			)}
		/>
	);
}
