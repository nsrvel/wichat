"use client";

import {
	createContext,
	useCallback,
	useContext,
	useEffect,
	useMemo,
	useState,
	type ReactNode,
} from "react";
type GroupSettingsUnsavedContextValue = {
	hasUnsavedChanges: boolean;
	setHasUnsavedChanges: (dirty: boolean) => void;
	runIfAllowed: (action: () => void) => void;
	footerNudgeKey: number;
};

const GroupSettingsUnsavedContext = createContext<GroupSettingsUnsavedContextValue | null>(null);

export function GroupSettingsUnsavedProvider({ children }: { children: ReactNode }) {
	const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
	const [footerNudgeKey, setFooterNudgeKey] = useState(0);

	const runIfAllowed = useCallback(
		(action: () => void) => {
			if (hasUnsavedChanges) {
				setFooterNudgeKey((key) => key + 1);
				return;
			}
			action();
		},
		[hasUnsavedChanges],
	);

	const value = useMemo(
		() => ({ hasUnsavedChanges, setHasUnsavedChanges, runIfAllowed, footerNudgeKey }),
		[footerNudgeKey, hasUnsavedChanges, runIfAllowed],
	);

	return (
		<GroupSettingsUnsavedContext.Provider value={value}>
			{children}
		</GroupSettingsUnsavedContext.Provider>
	);
}

export function useGroupSettingsUnsaved() {
	const ctx = useContext(GroupSettingsUnsavedContext);
	if (!ctx) {
		throw new Error("useGroupSettingsUnsaved must be used within GroupSettingsUnsavedProvider");
	}
	return ctx;
}

/** Profile (and future sections) report dirty state while mounted. */
export function useReportGroupSettingsUnsaved(dirty: boolean) {
	const { setHasUnsavedChanges } = useGroupSettingsUnsaved();
	useEffect(() => {
		setHasUnsavedChanges(dirty);
		return () => setHasUnsavedChanges(false);
	}, [dirty, setHasUnsavedChanges]);
}
