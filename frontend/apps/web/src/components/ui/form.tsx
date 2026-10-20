"use client";

import { cn } from "@/lib/utils";
import * as React from "react";
import {
	Controller,
	type ControllerProps,
	type FieldPath,
	type FieldValues,
	FormProvider,
	useFormContext,
} from "react-hook-form";
import { useTranslation } from "react-i18next";

const Form = FormProvider;

type FormFieldContextValue<
	TFieldValues extends FieldValues = FieldValues,
	TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
> = {
	name: TName;
};

const FormFieldContext = React.createContext<FormFieldContextValue>(
	{} as FormFieldContextValue,
);

const FormField = <
	TFieldValues extends FieldValues = FieldValues,
	TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
>({
	...props
}: ControllerProps<TFieldValues, TName>) => {
	return (
		<FormFieldContext.Provider value={{ name: props.name }}>
			<Controller {...props} />
		</FormFieldContext.Provider>
	);
};

const useFormField = () => {
	const fieldContext = React.useContext(FormFieldContext);
	const itemContext = React.useContext(FormItemContext);
	const { getFieldState, formState } = useFormContext();

	if (!fieldContext?.name) {
		throw new Error("useFormField should be used within <FormField>");
	}

	const fieldState = getFieldState(fieldContext.name, formState);
	const { id } = itemContext;

	return {
		id,
		name: fieldContext.name,
		formItemId: `${id}-form-item`,
		formDescriptionId: `${id}-form-item-description`,
		formMessageId: `${id}-form-item-message`,
		...fieldState,
	};
};

type FormItemContextValue = {
	id: string;
};

const FormItemContext = React.createContext<FormItemContextValue>({} as FormItemContextValue);

function FormItem({ className, ...props }: React.ComponentProps<"div">) {
	const id = React.useId();

	return (
		<FormItemContext.Provider value={{ id }}>
			<div className={cn("flex min-w-0 flex-col gap-1 sm:gap-1.5", className)} {...props} />
		</FormItemContext.Provider>
	);
}

function FormLabel({ className, ...props }: React.ComponentProps<"label">) {
	const { formItemId } = useFormField();

	return (
		<label htmlFor={formItemId} className={cn(className)} {...props} />
	);
}

function FormControl({ ...props }: React.ComponentProps<"div">) {
	const { error, formItemId, formDescriptionId, formMessageId } = useFormField();

	return (
		<div
			id={formItemId}
			aria-describedby={
				!error ? `${formDescriptionId}` : `${formDescriptionId} ${formMessageId}`
			}
			aria-invalid={!!error}
			{...props}
		/>
	);
}

function FormDescription({ className, ...props }: React.ComponentProps<"p">) {
	const { formDescriptionId } = useFormField();

	return (
		<p
			id={formDescriptionId}
			className={cn("text-[11px] text-muted-foreground", className)}
			{...props}
		/>
	);
}

function FormMessage({ className, children, ...props }: React.ComponentProps<"p">) {
	const { error, formMessageId } = useFormField();
	const { t } = useTranslation();
	const raw = error ? String(error.message ?? "") : children;

	if (!raw) {
		return null;
	}

	const body =
		typeof raw === "string" && raw.includes(".")
			? t(raw, { defaultValue: raw })
			: raw;

	return (
		<p
			id={formMessageId}
			className={cn("text-destructive text-[11px] font-medium", className)}
			{...props}
		>
			{body}
		</p>
	);
}

export {
	Form,
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
	useFormField,
};
