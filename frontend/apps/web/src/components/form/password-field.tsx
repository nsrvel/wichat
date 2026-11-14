"use client";

import { Eye, EyeOff } from "lucide-react";
import { type ComponentProps, type ReactNode, useId, useState } from "react";
import type { FieldPath, FieldValues } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { FieldLabelContent } from "@/components/form/field-label";
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { authFieldLabelClass, authInputClass } from "@/lib/form-field-styles";
import { cn } from "@/lib/utils";

type PasswordInputProps = Omit<ComponentProps<typeof Input>, "type"> & {
	showLabel?: string;
	hideLabel?: string;
};

export function PasswordInput({
	showLabel,
	hideLabel,
	className,
	...props
}: PasswordInputProps) {
	const { t } = useTranslation();
	const [visible, setVisible] = useState(false);
	const resolvedShowLabel = showLabel ?? t("login.show_password");
	const resolvedHideLabel = hideLabel ?? t("login.hide_password");

	return (
		<div className="relative">
			<Input
				type={visible ? "text" : "password"}
				className={cn(className, "pr-11 sm:pr-11")}
				{...props}
			/>
			<button
				type="button"
				className="text-muted-foreground hover:text-foreground absolute top-1/2 right-3 -translate-y-1/2"
				aria-label={visible ? resolvedHideLabel : resolvedShowLabel}
				onClick={() => setVisible((current) => !current)}
			>
				{visible ? <EyeOff className="size-4" aria-hidden /> : <Eye className="size-4" aria-hidden />}
			</button>
		</div>
	);
}

type PasswordFieldProps<T extends FieldValues> = {
	control: import("react-hook-form").Control<T>;
	name: FieldPath<T>;
	label: string;
	placeholder?: string;
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
	return (
		<FormField
			control={control}
			name={name}
			render={({ field, fieldState }) => (
				<FormItem>
					{labelAction ? (
						<div className="flex min-w-0 items-center justify-between gap-2">
							<FormLabel className={labelClassName}>
								<FieldLabelContent required={required}>{label}</FieldLabelContent>
							</FormLabel>
							{labelAction}
						</div>
					) : (
						<FormLabel className={labelClassName}>
							<FieldLabelContent required={required}>{label}</FieldLabelContent>
						</FormLabel>
					)}
					<FormControl>
						<PasswordInput
							autoComplete={autoComplete}
							placeholder={placeholder}
							aria-required={required}
							aria-invalid={!!fieldState.error}
							className={inputClassName}
							showLabel={showLabel}
							hideLabel={hideLabel}
							{...field}
						/>
					</FormControl>
					<FormMessage />
				</FormItem>
			)}
		/>
	);
}

type StandalonePasswordFieldProps = {
	id?: string;
	label: string;
	value: string;
	onChange: (value: string) => void;
	autoComplete?: "current-password" | "new-password";
	placeholder?: string;
	required?: boolean;
	error?: string | null;
	labelClassName?: string;
	inputClassName?: string;
	showLabel?: string;
	hideLabel?: string;
};

export function StandalonePasswordField({
	id: providedId,
	label,
	value,
	onChange,
	autoComplete = "current-password",
	placeholder,
	required = true,
	error,
	labelClassName,
	inputClassName,
	showLabel,
	hideLabel,
}: StandalonePasswordFieldProps) {
	const generatedId = useId();
	const id = providedId ?? generatedId;

	return (
		<div className="space-y-1.5">
			<label htmlFor={id} className={cn(authFieldLabelClass, "block", labelClassName)}>
				<FieldLabelContent required={required}>{label}</FieldLabelContent>
			</label>
			<PasswordInput
				id={id}
				value={value}
				onChange={(e) => onChange(e.target.value)}
				autoComplete={autoComplete}
				placeholder={placeholder}
				aria-required={required}
				aria-invalid={Boolean(error)}
				className={inputClassName ?? authInputClass}
				showLabel={showLabel}
				hideLabel={hideLabel}
			/>
			{error ? <p className="text-[11px] font-medium text-destructive">{error}</p> : null}
		</div>
	);
}
