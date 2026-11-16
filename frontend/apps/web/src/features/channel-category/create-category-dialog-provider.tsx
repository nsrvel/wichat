"use client";

import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import {
	CreateCategoryDialog,
	type CreateCategoryFormValues,
	type CreateCategoryTarget,
} from "@/features/channel-category/components/create-category-dialog";
import { translateError } from "@/integrations/i18n/translate-error";
import { useCreateGroupCategoryMutation } from "@/services/channel-category/channel-category.mutation";

type CreateCategoryDialogContextValue = {
	openCreateCategory: (target: CreateCategoryTarget) => void;
	closeCreateCategory: () => void;
};

const CreateCategoryDialogContext = createContext<CreateCategoryDialogContextValue | null>(null);

export function CreateCategoryDialogProvider({ children }: { children: ReactNode }) {
	const { t } = useTranslation();
	const [open, setOpen] = useState(false);
	const [target, setTarget] = useState<CreateCategoryTarget | null>(null);
	const createCategoryMutation = useCreateGroupCategoryMutation(target?.groupId ?? "");

	const openCreateCategory = useCallback((next: CreateCategoryTarget) => {
		setTarget(next);
		setOpen(true);
	}, []);

	const closeCreateCategory = useCallback(() => {
		setOpen(false);
		setTarget(null);
	}, []);

	const onOpenChange = useCallback((next: boolean) => {
		setOpen(next);
		if (!next) {
			setTarget(null);
		}
	}, []);

	const handleSubmitCategory = useCallback(
		(values: CreateCategoryFormValues, currentTarget: CreateCategoryTarget) => {
			if (!currentTarget.groupId) {
				return;
			}
			createCategoryMutation.mutate(
				{ name: values.name, is_private: values.isPrivate },
				{
					onSuccess: () => {
						closeCreateCategory();
					},
					onError: (error) => {
						toast.error(translateError(t, error));
					},
				},
			);
		},
		[closeCreateCategory, createCategoryMutation, t],
	);

	const value = useMemo(
		() => ({ openCreateCategory, closeCreateCategory }),
		[openCreateCategory, closeCreateCategory],
	);

	return (
		<CreateCategoryDialogContext.Provider value={value}>
			{children}
			<CreateCategoryDialog
				open={open}
				onOpenChange={onOpenChange}
				target={target}
				onSubmitCategory={handleSubmitCategory}
				submitting={createCategoryMutation.isPending}
			/>
		</CreateCategoryDialogContext.Provider>
	);
}

export function useCreateCategoryDialog() {
	const ctx = useContext(CreateCategoryDialogContext);
	if (!ctx) {
		throw new Error("useCreateCategoryDialog must be used within CreateCategoryDialogProvider");
	}
	return ctx;
}
