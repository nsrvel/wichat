"use client";

import { Moon, Sun } from "lucide-react";
import { useTheme } from "next-themes";
import { useTranslation } from "react-i18next";
import { useMounted } from "@/features/auth/use-mounted";

export function AuthThemeToggle() {
	const { t } = useTranslation();
	const mounted = useMounted();
	const { resolvedTheme, setTheme } = useTheme();

	if (!mounted) {
		return <span className="inline-block size-8 shrink-0" aria-hidden />;
	}

	const dark = resolvedTheme === "dark";

	return (
		<button
			type="button"
			onClick={() => setTheme(dark ? "light" : "dark")}
			className="text-muted-foreground hover:bg-muted inline-flex size-8 shrink-0 items-center justify-center rounded-lg"
			aria-label={t("theme.toggle")}
		>
			{dark ? (
				<Sun className="size-4" strokeWidth={1.8} />
			) : (
				<Moon className="size-4" strokeWidth={1.8} />
			)}
		</button>
	);
}
