import type { ReactNode } from "react";

export function FieldLabelContent({
	children,
	required,
}: {
	children: ReactNode;
	required?: boolean;
}) {
	return (
		<>
			{children}
			{required ? (
				<span className="text-destructive ml-0.5 font-medium" aria-hidden="true">
					*
				</span>
			) : null}
		</>
	);
}
