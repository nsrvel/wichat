import type { Metadata } from "next";
import { Geist_Mono } from "next/font/google";
import { cookies } from "next/headers";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";
import { LOCALE_COOKIE, localeFromCookie } from "@/integrations/i18n/locale";
import { AppProviders } from "@/integrations/tanstack-query/providers";
import { appSans } from "@/lib/fonts/app-sans";
import { I18nProvider } from "@wichat/i18n";
import "./globals.css";

const geistMono = Geist_Mono({
	variable: "--font-geist-mono",
	subsets: ["latin"],
});

export const metadata: Metadata = {
	title: "WiChat",
	description: "Self-hosted team communication",
	icons: {
		icon: [
			{ url: "/favicon/favicon.ico", sizes: "any" },
			{ url: "/favicon/favicon-32x32.png", sizes: "32x32", type: "image/png" },
			{ url: "/favicon/favicon-16x16.png", sizes: "16x16", type: "image/png" },
		],
		apple: [{ url: "/favicon/apple-touch-icon.png", sizes: "180x180" }],
	},
	manifest: "/favicon/site.webmanifest",
};

export default async function RootLayout({ children }: { children: React.ReactNode }) {
	const jar = await cookies();
	const locale = localeFromCookie(jar.get(LOCALE_COOKIE)?.value);

	return (
		<html
			lang={locale}
			suppressHydrationWarning
			className={`${appSans.variable} ${geistMono.variable} h-full antialiased`}
		>
			<body
				className={`${appSans.className} flex h-dvh min-h-0 flex-col overflow-hidden bg-background font-sans text-foreground`}
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
					<I18nProvider locale={locale}>
						<AppProviders>
							{children}
							<Toaster />
						</AppProviders>
					</I18nProvider>
				</ThemeProvider>
			</body>
		</html>
	);
}
