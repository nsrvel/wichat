import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { cookies } from "next/headers";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";
import { LOCALE_COOKIE, localeFromCookie } from "@/integrations/i18n/locale";
import { I18nProvider } from "@wichat/i18n";
import { AppProviders } from "@/integrations/tanstack-query/providers";
import "./globals.css";

const geistSans = Geist({
	variable: "--font-geist-sans",
	subsets: ["latin"],
});

const geistMono = Geist_Mono({
	variable: "--font-geist-mono",
	subsets: ["latin"],
});

export const metadata: Metadata = {
	title: "WiChat",
	description: "Self-hosted team communication",
};

export default async function RootLayout({ children }: { children: React.ReactNode }) {
	const jar = await cookies();
	const locale = localeFromCookie(jar.get(LOCALE_COOKIE)?.value);

	return (
		<html
			lang={locale}
			suppressHydrationWarning
			className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
		>
			<body
				className="flex h-dvh min-h-0 flex-col overflow-hidden"
				suppressHydrationWarning
			>
				<ThemeProvider attribute="class" defaultTheme="system" enableSystem>
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
