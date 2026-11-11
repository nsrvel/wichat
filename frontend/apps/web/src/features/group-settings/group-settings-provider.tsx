"use client";

import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import { GroupSettingsOverlay } from "@/features/group-settings/components/group-settings-overlay";
import type {
	GroupSettingsSection,
	GroupSettingsTarget,
	OpenGroupSettingsInput,
} from "@/features/group-settings/lib/group-settings-target";

type GroupSettingsContextValue = {
	openGroupSettings: (input: OpenGroupSettingsInput) => void;
	closeGroupSettings: () => void;
	patchGroupSettingsTarget: (patch: Partial<GroupSettingsTarget>) => void;
};

const GroupSettingsContext = createContext<GroupSettingsContextValue | null>(null);

export function GroupSettingsProvider({ children }: { children: ReactNode }) {
	const [open, setOpen] = useState(false);
	const [target, setTarget] = useState<GroupSettingsTarget | null>(null);
	const [section, setSection] = useState<GroupSettingsSection>("profile");

	const openGroupSettings = useCallback((input: OpenGroupSettingsInput) => {
		setTarget({
			groupId: input.groupId,
			groupName: input.groupName,
			logoUrl: input.logoUrl,
			description: input.description ?? "",
			bannerUrl: input.bannerUrl ?? "",
			bannerColor: input.bannerColor ?? "",
			isPrivate: input.isPrivate,
			role: input.role,
		});
		setSection(input.section ?? "profile");
		setOpen(true);
	}, []);

	const patchGroupSettingsTarget = useCallback((patch: Partial<GroupSettingsTarget>) => {
		setTarget((prev) => (prev ? { ...prev, ...patch } : prev));
	}, []);

	const closeGroupSettings = useCallback(() => {
		setOpen(false);
	}, []);

	const onOpenChange = useCallback(
		(next: boolean) => {
			if (!next) {
				closeGroupSettings();
			}
		},
		[closeGroupSettings],
	);

	const value = useMemo(
		() => ({ openGroupSettings, closeGroupSettings, patchGroupSettingsTarget }),
		[openGroupSettings, closeGroupSettings, patchGroupSettingsTarget],
	);

	return (
		<GroupSettingsContext.Provider value={value}>
			{children}
			<GroupSettingsOverlay
				open={open}
				onClose={() => onOpenChange(false)}
				target={target}
				section={section}
			/>
		</GroupSettingsContext.Provider>
	);
}

export function useGroupSettings() {
	const ctx = useContext(GroupSettingsContext);
	if (!ctx) {
		throw new Error("useGroupSettings must be used within GroupSettingsProvider");
	}
	return ctx;
}
