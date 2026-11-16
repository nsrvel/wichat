"use client";

import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import { ChannelSettingsOverlay } from "@/features/channel/channel-settings/channel-settings-overlay";
import type {
	ChannelSettingsSection,
	ChannelSettingsTarget,
} from "@/features/channel/channel-settings/channel-settings-sidebar";

type ChannelSettingsContextValue = {
	openChannelSettings: (
		target: ChannelSettingsTarget,
		initialSection?: ChannelSettingsSection,
	) => void;
	closeChannelSettings: () => void;
};

const ChannelSettingsContext = createContext<ChannelSettingsContextValue | null>(null);

export function ChannelSettingsProvider({ children }: { children: ReactNode }) {
	const [target, setTarget] = useState<ChannelSettingsTarget | null>(null);
	const [section, setSection] = useState<ChannelSettingsSection>("overview");

	const openChannelSettings = useCallback(
		(nextTarget: ChannelSettingsTarget, initialSection: ChannelSettingsSection = "overview") => {
			setTarget(nextTarget);
			setSection(initialSection);
		},
		[],
	);

	const closeChannelSettings = useCallback(() => {
		setTarget(null);
	}, []);

	const value = useMemo(
		() => ({ openChannelSettings, closeChannelSettings }),
		[openChannelSettings, closeChannelSettings],
	);

	return (
		<ChannelSettingsContext.Provider value={value}>
			{children}
			<ChannelSettingsOverlay
				open={target !== null}
				onClose={closeChannelSettings}
				target={target}
				section={section}
			/>
		</ChannelSettingsContext.Provider>
	);
}

export function useChannelSettings() {
	const ctx = useContext(ChannelSettingsContext);
	if (!ctx) {
		throw new Error("useChannelSettings must be used within ChannelSettingsProvider");
	}
	return ctx;
}
