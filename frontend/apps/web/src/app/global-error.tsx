"use client";

import { I18nProvider } from "@wichat/i18n";
import { Geist_Mono } from "next/font/google";
import { useEffect } from "react";
import { StatusScreen } from "@/components/status/status-screen";
import { ThemeProvider } from "@/components/theme-provider";
import { appSans } from "@/lib/fonts/app-sans";
import "./globals.css";

const geistMono = Geist_Mono({
	variable: "--font-geist-mono",
	subsets: ["latin"],
});

export default function GlobalErrorPage({
	error,
	reset,
}: {
	error: Error & { digest?: string };
	reset: () => void;
}) {
	useEffect(() => {
		console.error(error);
	}, [error]);

	return (
		<html
			lang="en"
			suppressHydrationWarning
			className={`${appSans.variable} ${geistMono.variable} dark h-full antialiased`}
		>
			<body
				className={`${appSans.className} flex min-h-dvh w-full flex-col bg-background font-sans text-foreground`}
				suppressHydrationWarning
			>
				<ThemeProvider
					attribute="class"
					defaultTheme="dark"
					forcedTheme="dark"
					enableSystem={false}
					storageKey="wichat-theme"
					disableTransitionOnChange
				>
					<I18nProvider locale="en">
						<StatusScreen variant="error" onRetry={reset} />
					</I18nProvider>
				</ThemeProvider>
			</body>
		</html>
	);
}
