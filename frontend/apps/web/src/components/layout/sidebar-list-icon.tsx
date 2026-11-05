import Image from "next/image";
import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

/** Shared 20×20 tile for sidebar rows (DM avatar, channel icon, quick views). */
export const sidebarListIconBoxClass =
	"flex h-5 w-5 shrink-0 items-center justify-center overflow-hidden rounded-full border border-[#ececec] bg-white shadow-[0_1px_1px_rgba(0,0,0,0.03)] dark:border-sidebar-border dark:bg-[#242424] lg:h-[22px] lg:w-[22px]";

/** Conversation header — same tile style, original 32×32 footprint. */
export const conversationHeaderIconBoxClass =
	"flex h-8 w-8 shrink-0 items-center justify-center overflow-hidden rounded-full border border-[#ececec] bg-white shadow-[0_1px_1px_rgba(0,0,0,0.03)] dark:border-sidebar-border dark:bg-[#242424]";

export type SidebarListIconVariant = "list" | "header";

function iconBoxClass(variant: SidebarListIconVariant) {
	return variant === "header" ? conversationHeaderIconBoxClass : sidebarListIconBoxClass;
}

export function SidebarListIconBox({
	children,
	className,
	variant = "list",
}: {
	children: ReactNode;
	className?: string;
	variant?: SidebarListIconVariant;
}) {
	return <span className={cn(iconBoxClass(variant), className)}>{children}</span>;
}

export function SidebarListAvatar({
	src,
	alt = "",
	variant = "list",
}: {
	src: string;
	alt?: string;
	variant?: SidebarListIconVariant;
}) {
	const px = variant === "header" ? 32 : 22;
	return (
		<SidebarListIconBox className="p-0" variant={variant}>
			<Image src={src} alt={alt} width={px} height={px} className="h-full w-full object-cover" />
		</SidebarListIconBox>
	);
}

/** 1:1 DM when no photo is available (same 20×20 tile as channels). */
export function SidebarListInitialsAvatar({
	initials,
	variant = "list",
}: {
	initials: string;
	variant?: SidebarListIconVariant;
}) {
	const text = initials.replace(/\s/g, "").slice(0, 2).toUpperCase();
	return (
		<SidebarListIconBox
			variant={variant}
			className={cn(
				"border-[#E4E4E7] bg-[#EEEEF0] p-0 font-semibold leading-none text-[#52525b]",
				variant === "header" ? "text-[10px]" : "text-[8px]",
			)}
		>
			{text}
		</SidebarListIconBox>
	);
}
