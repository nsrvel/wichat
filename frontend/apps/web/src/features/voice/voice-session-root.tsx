"use client";

import type { ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { sessionQueryOpts } from "@/services/auth/auth.query";
import { VoiceSessionProvider } from "@/features/voice/voice-session-provider";

export function VoiceSessionRoot({ children }: { children: ReactNode }) {
	const { data: user } = useQuery(sessionQueryOpts());
	return <VoiceSessionProvider currentUserId={user?.id}>{children}</VoiceSessionProvider>;
}
