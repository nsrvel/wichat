import { Crown } from "lucide-react";

type OwnerCrownIconProps = {
	label: string;
};

/**
 * Shared inline crown indicator displayed to the right of a group owner's name
 * in member lists.
 */
export function OwnerCrownIcon({ label }: OwnerCrownIconProps) {
	return (
		<span className="inline-flex shrink-0 items-center" title={label} aria-label={label}>
			<Crown
				className="size-3.5 text-amber-500 dark:text-amber-400"
				strokeWidth={2.2}
				aria-hidden
			/>
		</span>
	);
}
