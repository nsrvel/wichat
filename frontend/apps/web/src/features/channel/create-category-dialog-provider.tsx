"use client";

import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import {
	CreateCategoryDialog,
	type CreateCategoryTarget,
} from "@/features/channel/components/create-category-dialog";

type CreateCategoryDialogContextValue = {
	openCreateCategory: (target: CreateCategoryTarget) => void;
	closeCreateCategory: () => void;
};

const CreateCategoryDialogContext = createContext<CreateCategoryDialogContextValue | null>(null);

export function CreateCategoryDialogProvider({ children }: { children: ReactNode }) {
	const [open, setOpen] = useState(false);
	const [target, setTarget] = useState<CreateCategoryTarget | null>(null);

	const openCreateCategory = useCallback((next: CreateCategoryTarget) => {
		setTarget(next);
		setOpen(true);
	}, []);

	const closeCreateCategory = useCallback(() => {
		setOpen(false);
	}, []);

	const onOpenChange = useCallback((next: boolean) => {
		setOpen(next);
		if (!next) {
			setTarget(null);
		}
	}, []);

	const value = useMemo(
		() => ({ openCreateCategory, closeCreateCategory }),
		[openCreateCategory, closeCreateCategory],
	);

	return (
		<CreateCategoryDialogContext.Provider value={value}>
			{children}
			<CreateCategoryDialog open={open} onOpenChange={onOpenChange} target={target} />
		</CreateCategoryDialogContext.Provider>
	);
}

export function useCreateCategoryDialog() {
	const ctx = useContext(CreateCategoryDialogContext);
	if (!ctx) {
		throw new Error(
			"useCreateCategoryDialog must be used within CreateCategoryDialogProvider",
		);
	}
	return ctx;
}

