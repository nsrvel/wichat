"use client";

import { usePathname } from "next/navigation";
import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import { GroupOnboardingDialog } from "@/features/group/group-onboarding-flow";
import { useMyGroupsQuery } from "@/hooks/use-my-groups-query";
import { APP_HOME_PATH } from "@/lib/app-routes";

type GroupOnboardingContextValue = {
	open: boolean;
	openOnboarding: () => void;
	closeOnboarding: () => void;
};

const GroupOnboardingContext = createContext<GroupOnboardingContextValue | null>(null);

export function GroupOnboardingProvider({ children }: { children: ReactNode }) {
	const pathname = usePathname();
	const myGroups = useMyGroupsQuery();
	const [manualOpen, setManualOpen] = useState(false);
	const [dismissedOnHome, setDismissedOnHome] = useState(false);

	const onHome = pathname === APP_HOME_PATH;
	const noGroups = !myGroups.isPending && (myGroups.data ?? []).length === 0;
	const autoOpen = onHome && noGroups && !dismissedOnHome;
	const open = manualOpen || autoOpen;

	const openOnboarding = useCallback(() => {
		setDismissedOnHome(false);
		setManualOpen(true);
	}, []);
	const closeOnboarding = useCallback(() => setManualOpen(false), []);

	const setOpen = useCallback(
		(next: boolean) => {
			if (next) {
				setDismissedOnHome(false);
				setManualOpen(true);
				return;
			}
			setManualOpen(false);
			if (onHome && noGroups) {
				setDismissedOnHome(true);
			}
		},
		[noGroups, onHome],
	);

	const value = useMemo(
		() => ({ open, openOnboarding, closeOnboarding }),
		[open, openOnboarding, closeOnboarding],
	);

	return (
		<GroupOnboardingContext.Provider value={value}>
			{children}
			<GroupOnboardingDialog open={open} onOpenChange={setOpen} />
		</GroupOnboardingContext.Provider>
	);
}

export function useGroupOnboarding() {
	const ctx = useContext(GroupOnboardingContext);
	if (!ctx) {
		throw new Error("useGroupOnboarding must be used within GroupOnboardingProvider");
	}
	return ctx;
}
