"use client";

import { useId } from "react";
import { Radio } from "lucide-react";
import { useTranslation } from "react-i18next";
import {
	WiChatDialog,
	WiChatDialogBody,
	WiChatDialogFooter,
	WiChatDialogFooterActions,
	WiChatDialogHeader,
} from "@/components/ui/wichat-dialog";

interface VoiceRecordingConsentDialogProps {
	open: boolean;
	onAccept: () => void;
	onLeave: () => void;
}

export function VoiceRecordingConsentDialog({
	open,
	onAccept,
	onLeave,
}: VoiceRecordingConsentDialogProps) {
	const { t } = useTranslation();
	const titleId = useId();

	return (
		<WiChatDialog
			open={open}
			onOpenChange={(next) => {
				if (!next) onLeave();
			}}
			titleId={titleId}
		>
			<WiChatDialogHeader
				titleId={titleId}
				title={
					<span className="flex items-center gap-2">
						<Radio className="size-4 animate-pulse text-red-500" />
						{t("voice.recording.consent_title")}
					</span>
				}
			/>
			<WiChatDialogBody>
				<p className="text-sm text-text-secondary">{t("voice.recording.consent_description")}</p>
			</WiChatDialogBody>
			<WiChatDialogFooter>
				<WiChatDialogFooterActions
					backLabel={t("voice.recording.consent_leave")}
					onBack={onLeave}
					primaryLabel={t("voice.recording.consent_accept")}
					onPrimary={onAccept}
				/>
			</WiChatDialogFooter>
		</WiChatDialog>
	);
}
