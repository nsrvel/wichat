import { cn } from "cn";
import type { ComponentProps } from "react";

function Input({ className, type = "text", ...props }: ComponentProps<"input">) {
	return (
		<input
			type={type}
			data-slot="input"
			className={cn(
				"border-input bg-background text-foreground placeholder:text-muted-foreground h-9 w-full rounded-lg border px-3 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:opacity-50",
				className,
			)}
			{...props}
		/>
	);
}

export { Input };
