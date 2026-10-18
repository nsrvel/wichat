import { cn } from "cn";
import type { ComponentProps } from "react";

function Label({ className, ...props }: ComponentProps<"label">) {
	return (
		<label
			data-slot="label"
			className={cn("text-foreground text-sm font-medium", className)}
			{...props}
		/>
	);
}

export { Label };
