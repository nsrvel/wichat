import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import type { GroupOccupancyResponse } from "@wichat/contracts/chat/voice";
import { subscribeVoiceRealtime } from "@/integrations/realtime/voice-realtime";
import { Q_VOICE_OCCUPANCY, Q_VOICE_RECORDINGS } from "@/services/voice/voice.key";

/** Keeps group voice occupancy and recording state in sync with gateway WS events. */
export function useVoiceOccupancyRealtime(groupId?: string) {
	const qc = useQueryClient();

	useEffect(() => {
		if (!groupId) return;
		return subscribeVoiceRealtime((ev) => {
			if (ev.group_id !== groupId) return;

			if (ev.type === "voice.occupancy.updated") {
				qc.setQueryData<GroupOccupancyResponse>([Q_VOICE_OCCUPANCY, groupId], (prev) => {
					const existing = prev?.channels ?? [];
					const others = existing.filter((c) => c.channel_id !== ev.channel_id);
					if (ev.occupants.length === 0 && !ev.is_recording) {
						return { channels: others };
					}
					return {
						channels: [
							...others,
							{
								channel_id: ev.channel_id,
								is_recording: ev.is_recording,
								occupants: ev.occupants,
							},
						],
					};
				});
			} else if (ev.type === "voice.recording.updated") {
				qc.setQueryData<GroupOccupancyResponse>([Q_VOICE_OCCUPANCY, groupId], (prev) => {
					if (!prev) return prev;
					return {
						channels: prev.channels.map((c) =>
							c.channel_id === ev.channel_id ? { ...c, is_recording: ev.is_recording } : c,
						),
					};
				});
				void qc.invalidateQueries({ queryKey: [Q_VOICE_RECORDINGS, ev.channel_id] });
			}
		});
	}, [groupId, qc]);
}
