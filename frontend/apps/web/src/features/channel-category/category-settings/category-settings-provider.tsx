"use client";

import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import { CategorySettingsOverlay } from "@/features/channel-category/category-settings/category-settings-overlay";
import type {
	CategorySettingsSection,
	CategorySettingsTarget,
} from "@/features/channel-category/category-settings/category-settings-sidebar";

type CategorySettingsContextValue = {
	openCategorySettings: (
		target: CategorySettingsTarget,
		initialSection?: CategorySettingsSection,
	) => void;
	closeCategorySettings: () => void;
};

const CategorySettingsContext = createContext<CategorySettingsContextValue | null>(null);

export function CategorySettingsProvider({ children }: { children: ReactNode }) {
	const [target, setTarget] = useState<CategorySettingsTarget | null>(null);
	const [section, setSection] = useState<CategorySettingsSection>("overview");

	const openCategorySettings = useCallback(
		(nextTarget: CategorySettingsTarget, initialSection: CategorySettingsSection = "overview") => {
			setTarget(nextTarget);
			setSection(initialSection);
		},
		[],
	);

	const closeCategorySettings = useCallback(() => {
		setTarget(null);
	}, []);

	const value = useMemo(
		() => ({ openCategorySettings, closeCategorySettings }),
		[openCategorySettings, closeCategorySettings],
	);

	return (
		<CategorySettingsContext.Provider value={value}>
			{children}
			<CategorySettingsOverlay
				open={target !== null}
				onClose={closeCategorySettings}
				target={target}
				section={section}
			/>
		</CategorySettingsContext.Provider>
	);
}

export function useCategorySettings() {
	const ctx = useContext(CategorySettingsContext);
	if (!ctx) {
		throw new Error("useCategorySettings must be used within CategorySettingsProvider");
	}
	return ctx;
}
