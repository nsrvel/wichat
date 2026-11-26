import type { Popover as PopoverPrimitive } from "@base-ui/react/popover";
import type { ComponentProps } from "react";

type PopoverPositionerProps = ComponentProps<typeof PopoverPrimitive.Positioner>;

/** Boundary for flip/shift — viewport, not only `overflow-hidden` ancestors. */
export function floatingCollisionBoundary(): Element | "clipping-ancestors" {
	if (typeof document === "undefined") {
		return "clipping-ancestors";
	}
	return document.documentElement;
}

/**
 * Collision avoidance (Floating UI): flip to the opposite side when tight,
 * shift on the align axis when flip is not enough.
 */
export const floatingCollisionAvoidance = {
	side: "flip",
	align: "shift",
	fallbackAxisSide: "end",
} as const satisfies PopoverPositionerProps["collisionAvoidance"];

export const floatingCollisionPadding = 8;

/** Spread onto Base UI `*.Positioner` (Popover, Menu, ContextMenu, Tooltip). */
export function floatingPositionerDefaults(): Pick<
	PopoverPositionerProps,
	"collisionBoundary" | "collisionPadding" | "collisionAvoidance"
> {
	return {
		collisionBoundary: floatingCollisionBoundary(),
		collisionPadding: floatingCollisionPadding,
		collisionAvoidance: floatingCollisionAvoidance,
	};
}