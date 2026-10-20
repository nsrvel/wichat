"use client";

import { FieldLabelContent } from "@/components/form/field-label";
import { CustomSelect } from "@/components/ui/custom-select";
import {
	FormControl,
	FormField,
	FormItem,
	FormMessage,
} from "@/components/ui/form";
import { birthDateMonthOptions } from "@/lib/birth-date-options";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import type { FieldPath, FieldValues } from "react-hook-form";

const currentYear = new Date().getFullYear();
const YEARS = Array.from({ length: 100 }, (_, i) => {
	const y = String(currentYear - i);
	return { value: y, label: y };
});
const DAYS = Array.from({ length: 31 }, (_, i) => {
	const d = String(i + 1).padStart(2, "0");
	return { value: d, label: d };
});

type BirthDateFieldProps<T extends FieldValues> = {
	control: import("react-hook-form").Control<T>;
	dayName: FieldPath<T>;
	monthName: FieldPath<T>;
	yearName: FieldPath<T>;
	label: string;
	dayPlaceholder: string;
	monthPlaceholder: string;
	yearPlaceholder: string;
	labelClassName?: string;
	selectClassName: string;
	required?: boolean;
};

export function BirthDateField<T extends FieldValues>({
	control,
	dayName,
	monthName,
	yearName,
	label,
	dayPlaceholder,
	monthPlaceholder,
	yearPlaceholder,
	labelClassName,
	selectClassName,
	required = false,
}: BirthDateFieldProps<T>) {
	const { i18n } = useTranslation();
	const months = useMemo(() => birthDateMonthOptions(i18n.language), [i18n.language]);

	return (
		<div className="@container flex min-w-0 flex-col gap-1 sm:gap-1.5">
			<span className={labelClassName}>
				<FieldLabelContent required={required}>{label}</FieldLabelContent>
			</span>
			<div
				className="grid min-w-0 grid-cols-[minmax(0,0.8fr)_minmax(0,1.4fr)_minmax(0,0.95fr)] gap-1 sm:gap-2"
			>
				<FormField
					control={control}
					name={dayName}
					render={({ field }) => (
						<FormItem className="gap-0">
							<FormControl>
								<CustomSelect
									options={DAYS}
									value={field.value}
									placeholder={dayPlaceholder}
									onChange={field.onChange}
									triggerClassName={selectClassName}
									typeahead
								/>
							</FormControl>
						</FormItem>
					)}
				/>
				<FormField
					control={control}
					name={monthName}
					render={({ field }) => (
						<FormItem className="gap-0">
							<FormControl>
								<CustomSelect
									options={months}
									value={field.value}
									placeholder={monthPlaceholder}
									onChange={field.onChange}
									triggerClassName={selectClassName}
									typeahead
								/>
							</FormControl>
						</FormItem>
					)}
				/>
				<FormField
					control={control}
					name={yearName}
					render={({ field }) => (
						<FormItem className="gap-0">
							<FormControl>
								<CustomSelect
									options={YEARS}
									value={field.value}
									placeholder={yearPlaceholder}
									onChange={field.onChange}
									triggerClassName={selectClassName}
									typeahead
								/>
							</FormControl>
							<FormMessage />
						</FormItem>
					)}
				/>
			</div>
		</div>
	);
}
