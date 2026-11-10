import type { ReactNode } from "react";
import { authFieldLabelClass } from "@/features/auth/components/auth-shell";
import { cn } from "@/lib/utils";

export function ShareDialogSection({
	label,
	children,
	className,
}: {
	label: ReactNode;
	children: ReactNode;
	className?: string;
}) {
	return (
		<section className={cn("space-y-2", className)}>
			<h3 className={authFieldLabelClass}>{label}</h3>
			{children}
		</section>
	);
}
