"use client";

import { usePathname } from "next/navigation";
import { isUiSampleChatPath } from "@/lib/ui-sample-route";

export function useUiSampleChatRoute() {
	const pathname = usePathname();
	return isUiSampleChatPath(pathname);
}
