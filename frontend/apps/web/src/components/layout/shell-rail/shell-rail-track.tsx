"use client";

import type { ReactNode } from "react";
import {
	shellRailColumnClass,
	shellRailTrackClass,
} from "@/components/layout/shell-rail/shell-rail-aside";
type ShellRailTrackProps = {
	open: boolean;
	rail: ReactNode;
	children: ReactNode;
	className?: string;
};

/** Rail column width `w-0` ↔ `w-16`; motion tokens in `globals.css` `:root`. */
export function ShellRailTrack({ open, rail, children, className }: ShellRailTrackProps) {
	return (
		<div className={shellRailTrackClass(className)}>
			<div className={shellRailColumnClass(open)}>{rail}</div>
			{children}
		</div>
	);
}
