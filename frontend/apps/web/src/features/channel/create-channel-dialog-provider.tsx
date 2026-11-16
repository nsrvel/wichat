"use client";

import { useRouter } from "next/navigation";
import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import {
	CreateChannelDialog,
	type CreateChannelFormValues,
	type CreateChannelTarget,
} from "@/features/channel/components/create-channel-dialog";
import { translateError } from "@/integrations/i18n/translate-error";
import { channelPath } from "@/lib/app-routes";
import { useCreateGroupChannelMutation } from "@/services/channel/channel.mutation";

type CreateChannelDialogContextValue = {
	openCreateChannel: (target: CreateChannelTarget) => void;
	closeCreateChannel: () => void;
};

const CreateChannelDialogContext = createContext<CreateChannelDialogContextValue | null>(null);

export function CreateChannelDialogProvider({ children }: { children: ReactNode }) {
	const { t } = useTranslation();
	const router = useRouter();
	const [open, setOpen] = useState(false);
	const [target, setTarget] = useState<CreateChannelTarget | null>(null);
	const createChannelMutation = useCreateGroupChannelMutation(target?.groupId ?? "");

	const openCreateChannel = useCallback((next: CreateChannelTarget) => {
		setTarget(next);
		setOpen(true);
	}, []);

	const closeCreateChannel = useCallback(() => {
		setOpen(false);
		setTarget(null);
	}, []);

	const onOpenChange = useCallback((next: boolean) => {
		setOpen(next);
		if (!next) {
			setTarget(null);
		}
	}, []);

	const handleSubmitChannel = useCallback(
		(values: CreateChannelFormValues, currentTarget: CreateChannelTarget) => {
			if (!currentTarget.groupId) {
				return;
			}
			createChannelMutation.mutate(
				{
					name: values.name,
					type: values.type,
					category_id: values.categoryId || undefined,
					is_private: values.isPrivate,
				},
				{
					onSuccess: (created) => {
						closeCreateChannel();
						router.push(channelPath(currentTarget.groupId, created.id));
					},
					onError: (error) => {
						toast.error(translateError(t, error));
					},
				},
			);
		},
		[closeCreateChannel, createChannelMutation, router, t],
	);

	const value = useMemo(
		() => ({ openCreateChannel, closeCreateChannel }),
		[openCreateChannel, closeCreateChannel],
	);

	return (
		<CreateChannelDialogContext.Provider value={value}>
			{children}
			<CreateChannelDialog
				open={open}
				onOpenChange={onOpenChange}
				target={target}
				onSubmitChannel={handleSubmitChannel}
				submitting={createChannelMutation.isPending}
			/>
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
