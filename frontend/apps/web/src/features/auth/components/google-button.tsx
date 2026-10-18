"use client";

import { Button } from "@/components/ui/button";
import { googleStartUrl } from "@/services/auth/auth.service";

export function GoogleButton({ label }: { label: string }) {
	return (
		<Button
			type="button"
			variant="outline"
			className="h-11 w-full rounded-lg border-black/10 bg-transparent text-sm font-medium text-[#171717] shadow-none hover:bg-black/5 dark:border-white/15 dark:bg-transparent dark:text-white dark:hover:bg-white/5"
			onClick={() => {
				window.location.assign(googleStartUrl());
			}}
		>
			{label}
		</Button>
	);
}
