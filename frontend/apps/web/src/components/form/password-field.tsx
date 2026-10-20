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
import { Eye, EyeOff } from "lucide-react";
import { useState } from "react";
import type { ReactNode } from "react";
import type { FieldPath, FieldValues } from "react-hook-form";

type PasswordFieldProps<T extends FieldValues> = {
	control: import("react-hook-form").Control<T>;
	name: FieldPath<T>;
	label: string;
	placeholder: string;
	autoComplete: "current-password" | "new-password";
	showLabel: string;
	hideLabel: string;
	labelClassName?: string;
	inputClassName?: string;
	required?: boolean;
	labelAction?: ReactNode;
};

export function PasswordField<T extends FieldValues>({
	control,
	name,
	label,
	placeholder,
	autoComplete,
	showLabel,
	hideLabel,
	labelClassName,
	inputClassName,
	required = false,
	labelAction,
}: PasswordFieldProps<T>) {
	const [visible, setVisible] = useState(false);

	return (
		<FormField
			control={control}
			name={name}
			render={({ field, fieldState }) => (
				<FormItem>
					{labelAction ? (
						<div className="flex min-w-0 items-center justify-between gap-2">
							<span className={labelClassName}>
								<FieldLabelContent required={required}>{label}</FieldLabelContent>
							</span>
							{labelAction}
						</div>
					) : (
						<FormLabel className={labelClassName}>
							<FieldLabelContent required={required}>{label}</FieldLabelContent>
						</FormLabel>
					)}
					<FormControl>
						<div className="relative">
							<Input
								type={visible ? "text" : "password"}
								autoComplete={autoComplete}
								placeholder={placeholder}
								aria-required={required}
								aria-invalid={!!fieldState.error}
								className={`${inputClassName ?? ""} pr-10`}
								{...field}
							/>
							<button
								type="button"
								className="text-muted-foreground hover:text-foreground absolute top-1/2 right-3 -translate-y-1/2"
								aria-label={visible ? hideLabel : showLabel}
								onClick={() => setVisible((current) => !current)}
							>
								{visible ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
							</button>
						</div>
					</FormControl>
					<FormMessage />
				</FormItem>
			)}
		/>
	);
}
