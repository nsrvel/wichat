import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { ModerateParticipantRequest } from "@wichat/contracts/chat/voice";
import { Q_VOICE_OCCUPANCY, Q_VOICE_RECORDINGS } from "@/services/voice/voice.key";
import {
	acknowledgeVoiceIdleWarning,
	moderateVoiceParticipant,
	startCallRecording,
	stopCallRecording,
} from "@/services/voice/voice.service";

export function useModerateVoiceParticipantMutation(channelId: string, groupId?: string) {
	const qc = useQueryClient();
	return useMutation({
		mutationFn: (payload: ModerateParticipantRequest) =>
			moderateVoiceParticipant(channelId, payload),
		onSuccess: () => {
			if (groupId) {
				void qc.invalidateQueries({ queryKey: [Q_VOICE_OCCUPANCY, groupId] });
			}
		},
	});
}

export function useAcknowledgeVoiceIdleMutation(channelId: string, groupId?: string) {
	return useMutation({
		mutationFn: () => acknowledgeVoiceIdleWarning(channelId, groupId),
	});
}

export function useStartRecordingMutation(channelId: string, groupId?: string) {
	const qc = useQueryClient();
	return useMutation({
		mutationFn: () => startCallRecording(channelId, groupId),
		onSuccess: () => {
			if (groupId) {
				void qc.invalidateQueries({ queryKey: [Q_VOICE_OCCUPANCY, groupId] });
			}
			void qc.invalidateQueries({ queryKey: [Q_VOICE_RECORDINGS, channelId] });
		},
	});
}

export function useStopRecordingMutation(channelId: string, groupId?: string) {
	const qc = useQueryClient();
	return useMutation({
		mutationFn: (recordingId?: string) => stopCallRecording(channelId, groupId, recordingId),
		onSuccess: () => {
			if (groupId) {
				void qc.invalidateQueries({ queryKey: [Q_VOICE_OCCUPANCY, groupId] });
			}
			void qc.invalidateQueries({ queryKey: [Q_VOICE_RECORDINGS, channelId] });
		},
	});
}
