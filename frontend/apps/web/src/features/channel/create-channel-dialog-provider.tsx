"use client";

import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import {
	CreateChannelDialog,
	type CreateChannelTarget,
} from "@/features/channel/components/create-channel-dialog";

type CreateChannelDialogContextValue = {
	openCreateChannel: (target: CreateChannelTarget) => void;
	closeCreateChannel: () => void;
};

const CreateChannelDialogContext = createContext<CreateChannelDialogContextValue | null>(null);

export function CreateChannelDialogProvider({ children }: { children: ReactNode }) {
	const [open, setOpen] = useState(false);
	const [target, setTarget] = useState<CreateChannelTarget | null>(null);

	const openCreateChannel = useCallback((next: CreateChannelTarget) => {
		setTarget(next);
		setOpen(true);
	}, []);

	const closeCreateChannel = useCallback(() => {
		setOpen(false);
	}, []);

	const onOpenChange = useCallback((next: boolean) => {
		setOpen(next);
		if (!next) {
			setTarget(null);
		}
	}, []);

	const value = useMemo(
		() => ({ openCreateChannel, closeCreateChannel }),
		[openCreateChannel, closeCreateChannel],
	);

	return (
		<CreateChannelDialogContext.Provider value={value}>
			{children}
			<CreateChannelDialog open={open} onOpenChange={onOpenChange} target={target} />
		</CreateChannelDialogContext.Provider>
	);
}

export function useCreateChannelDialog() {
	const ctx = useContext(CreateChannelDialogContext);
	if (!ctx) {
		throw new Error("useCreateChannelDialog must be used within CreateChannelDialogProvider");
	}
	return ctx;
}
