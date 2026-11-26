import { createEnv } from "@t3-oss/env-nextjs";
import { z } from "zod";

export const env = createEnv({
	client: {
		NEXT_PUBLIC_API_URL: z.url(),
		NEXT_PUBLIC_REALTIME_WS_URL: z.string().url().optional(),
		NEXT_PUBLIC_OBJECT_PUBLIC_BASE_URL: z.string().url().optional(),
		NEXT_PUBLIC_MFA_ENABLED: z.enum(["true", "false"]).optional(),
	},
	runtimeEnv: {
		NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
		NEXT_PUBLIC_REALTIME_WS_URL: process.env.NEXT_PUBLIC_REALTIME_WS_URL,
		NEXT_PUBLIC_OBJECT_PUBLIC_BASE_URL: process.env.NEXT_PUBLIC_OBJECT_PUBLIC_BASE_URL,
		NEXT_PUBLIC_MFA_ENABLED: process.env.NEXT_PUBLIC_MFA_ENABLED,
	},
	emptyStringAsUndefined: true,
});
