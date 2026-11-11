"use client";

import { useQueryClient } from "@tanstack/react-query";
import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import { InviteDialog } from "@/features/invite/components/invite-dialog";
import type { InviteTarget } from "@/features/invite/lib/invite-target";
import { groupMembersQueryOpts } from "@/services/invite/invite.query";

type InviteDialogContextValue = {
	openInvite: (target: InviteTarget) => void;
	closeInvite: () => void;
};

const InviteDialogContext = createContext<InviteDialogContextValue | null>(null);

export function InviteDialogProvider({ children }: { children: ReactNode }) {
	const queryClient = useQueryClient();
	const [open, setOpen] = useState(false);
	const [target, setTarget] = useState<InviteTarget | null>(null);

	const openInvite = useCallback(
		(next: InviteTarget) => {
			setTarget(next);
			setOpen(true);
			void queryClient.prefetchQuery(groupMembersQueryOpts(next.groupId));
		},
		[queryClient],
	);

	const closeInvite = useCallback(() => {
		setOpen(false);
	}, []);

	const onOpenChange = useCallback((next: boolean) => {
		setOpen(next);
		if (!next) {
			setTarget(null);
		}
	}, []);

	const value = useMemo(() => ({ openInvite, closeInvite }), [openInvite, closeInvite]);

	return (
		<InviteDialogContext.Provider value={value}>
			{children}
			<InviteDialog open={open} onOpenChange={onOpenChange} target={target} />
		</InviteDialogContext.Provider>
	);
}

export function useInviteDialog() {
	const ctx = useContext(InviteDialogContext);
	if (!ctx) {
		throw new Error("useInviteDialog must be used within InviteDialogProvider");
	}
	return ctx;
}
