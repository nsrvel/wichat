"use client";

import { ContextMenu } from "@base-ui/react/context-menu";
import { Menu } from "@base-ui/react/menu";
import type { ComponentProps } from "react";
import { floatingPositionerDefaults } from "@/lib/floating-position";
import { menuPositionerClass } from "@/lib/menu-chrome";
import { cn } from "@/lib/utils";

type MenuPositionerProps = ComponentProps<typeof Menu.Positioner>;
type ContextMenuPositionerProps = ComponentProps<typeof ContextMenu.Positioner>;

const floating = floatingPositionerDefaults();

export function WiChatMenuPositioner({ className, ...props }: MenuPositionerProps) {
	return (
		<Menu.Positioner
			{...floating}
			{...props}
			className={cn(menuPositionerClass, className)}
		/>
	);
}

export function WiChatContextMenuPositioner({
	className,
	...props
}: ContextMenuPositionerProps) {
	return (
		<ContextMenu.Positioner
			{...floating}
			{...props}
			className={cn(menuPositionerClass, className)}
		/>
	);
}