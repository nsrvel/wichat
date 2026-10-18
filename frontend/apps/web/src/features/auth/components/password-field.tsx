"use client";

import { Eye, EyeOff } from "lucide-react";
import { useState } from "react";
import type { UseFormRegisterReturn } from "react-hook-form";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { authInputClass } from "@/features/auth/components/auth-shell";

export function PasswordField({
	id,
	label,
	placeholder,
	autoComplete,
	registration,
	error,
	showLabel,
	hideLabel,
}: {
	id: string;
	label: string;
	placeholder: string;
	autoComplete: "current-password" | "new-password";
	registration: UseFormRegisterReturn;
	error?: string;
	showLabel: string;
	hideLabel: string;
}) {
	const [visible, setVisible] = useState(false);

	return (
		<div className="flex flex-col gap-1.5">
			<Label htmlFor={id} className="sr-only">
				{label}
			</Label>
			<div className="relative">
				<Input
					id={id}
					type={visible ? "text" : "password"}
					autoComplete={autoComplete}
					placeholder={placeholder}
					className={`${authInputClass} pr-10`}
					{...registration}
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
			{error ? <p className="text-destructive text-[11px] font-medium">{error}</p> : null}
		</div>
	);
}
