"use client";

import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";

export type ShellNavKind = "dm" | "channel";

export type ShellNavTarget = {
	id: string;
	label: string;
	kind: ShellNavKind;
};

const DEFAULT_TARGET: ShellNavTarget = {
	id: "ch-wichat",
	label: "WiChat",
	kind: "channel",
};

type ShellNavigationContextValue = {
	active: ShellNavTarget;
	navigate: (target: ShellNavTarget) => void;
	isActive: (id: string) => boolean;
};

const ShellNavigationContext = createContext<ShellNavigationContextValue | null>(null);

export function ShellNavigationProvider({ children }: { children: ReactNode }) {
	const [active, setActive] = useState<ShellNavTarget>(DEFAULT_TARGET);

	const navigate = useCallback((target: ShellNavTarget) => {
		setActive((current) => (current.id === target.id ? current : target));
	}, []);

	const value = useMemo(
		() => ({
			active,
			navigate,
			isActive: (id: string) => active.id === id,
		}),
		[active, navigate],
	);

	return (
		<ShellNavigationContext.Provider value={value}>{children}</ShellNavigationContext.Provider>
	);
}

export function useShellNavigation() {
	const ctx = useContext(ShellNavigationContext);
	if (!ctx) {
		throw new Error("useShellNavigation must be used within ShellNavigationProvider");
	}
	return ctx;
}
