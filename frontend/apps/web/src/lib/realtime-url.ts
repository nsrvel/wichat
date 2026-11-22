import { env } from "@/lib/env";

/** HTTP(S) URL for the realtime upgrade (converted to ws/wss by the caller). */
export function realtimeHttpUrl(): string {
	const override = env.NEXT_PUBLIC_REALTIME_WS_URL?.replace(/\/+$/, "");
	if (override) {
		return override.replace(/^ws/, "http");
	}

	const base = env.NEXT_PUBLIC_API_URL.replace(/\/+$/, "");
	return base.endsWith("/api/v1") ? `${base}/realtime` : `${base}/api/v1/realtime`;
}