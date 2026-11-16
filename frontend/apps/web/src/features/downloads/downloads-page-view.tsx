"use client";

import { Download, Monitor, Smartphone } from "lucide-react";
import { useSyncExternalStore } from "react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { BrandHeader } from "@/components/layout/brand-header";
import { Button } from "@/components/ui/button";

type DesktopOs = "macos" | "windows" | "linux";

const emptySubscribe = () => () => {};

function detectDesktopOs(): DesktopOs {
	if (typeof window === "undefined") {
		return "macos";
	}
	const ua = window.navigator.userAgent;
	if (/Win/i.test(ua)) {
		return "windows";
	}
	if (/Linux/i.test(ua) && !/Android/i.test(ua)) {
		return "linux";
	}
	return "macos";
}

function AppleMark() {
	return (
		<svg viewBox="0 0 24 24" fill="currentColor" className="size-6 shrink-0" aria-hidden>
			<path d="M16.365 1.43c0 1.14-.493 2.27-1.177 3.08-.744.9-1.99 1.57-2.987 1.57-.12 0-.23-.02-.3-.03-.01-.06-.04-.22-.04-.39 0-1.15.572-2.27 1.206-2.98.804-.94 2.142-1.64 3.248-1.68.03.13.05.28.05.43zm4.565 15.71c-.03.07-.463 1.58-1.518 3.12-.945 1.34-1.94 2.71-3.43 2.71-1.517 0-1.9-.88-3.63-.88-1.698 0-2.302.91-3.67.91-1.377 0-2.332-1.26-3.428-2.8-1.287-1.82-2.323-4.63-2.323-7.28 0-4.28 2.797-6.55 5.552-6.55 1.448 0 2.675.95 3.6.95.865 0 2.222-1.01 3.902-1.01.613 0 2.886.06 4.374 2.19-.13.09-2.383 1.37-2.383 4.19 0 3.26 2.854 4.42 2.955 4.45z" />
		</svg>
	);
}

function GooglePlayMark() {
	return (
		<svg viewBox="0 0 24 24" className="size-6 shrink-0" aria-hidden>
			<path
				fill="#4285F4"
				d="M3.61 1.814L13.793 12 3.61 22.186a1.75 1.75 0 0 1-.46-1.19V3.004c0-.455.174-.883.46-1.19z"
			/>
			<path
				fill="#34A853"
				d="M17.207 8.586L13.793 12 3.61 1.814c.39-.245.887-.277 1.32-.045l12.277 6.817z"
			/>
			<path
				fill="#FBBC05"
				d="M20.84 10.604c1.09.605 1.09 2.187 0 2.792l-3.633 2.018L13.793 12l3.414-3.414 3.633 2.018z"
			/>
			<path
				fill="#EA4335"
				d="M13.793 12l3.414 3.414-12.277 6.817c-.433.232-.93.2-1.32-.045L13.793 12z"
			/>
		</svg>
	);
}

export function DownloadsPageView() {
	const { t } = useTranslation();
	const desktopOs = useSyncExternalStore(emptySubscribe, detectDesktopOs, (): DesktopOs => "macos");

	const desktopCtaLabel =
		desktopOs === "windows"
			? t("downloads_page.desktop_cta_windows")
			: desktopOs === "linux"
				? t("downloads_page.desktop_cta_linux")
				: t("downloads_page.desktop_cta_macos");

	const handleComingSoon = () => {
		toast.info(t("downloads_page.coming_soon_toast"));
	};

	return (
		<main className="overlay-scroll flex h-dvh min-h-0 flex-col overflow-x-hidden overflow-y-auto bg-background pt-[max(0.75rem,env(safe-area-inset-top))] pr-[max(1rem,env(safe-area-inset-right))] pb-[max(1.25rem,env(safe-area-inset-bottom))] pl-[max(1rem,env(safe-area-inset-left))] text-foreground sm:px-6 sm:pt-6 sm:pb-8 md:px-10 md:pb-10">
			<BrandHeader />
			<div className="mx-auto flex w-full max-w-6xl flex-1 flex-col justify-center px-2 py-12 sm:px-4 sm:py-16">
				{/* Row 1: Desktop Showcase */}
				<section className="grid items-center gap-10 lg:grid-cols-2 lg:gap-20">
					<div className="relative aspect-[16/10] w-full overflow-hidden rounded-3xl border border-border/60 bg-gradient-to-br from-sky-500/20 via-blue-500/15 to-indigo-500/20 p-6 sm:p-8">
						<div className="flex h-full w-full flex-col overflow-hidden rounded-2xl border border-border/70 bg-card/90 shadow-2xl backdrop-blur-sm">
							<div className="flex items-center gap-1.5 border-b border-border/60 px-4 py-2.5">
								<span className="size-2.5 rounded-full bg-rose-400/80" />
								<span className="size-2.5 rounded-full bg-amber-400/80" />
								<span className="size-2.5 rounded-full bg-emerald-400/80" />
							</div>
							<div className="flex flex-1 flex-col items-center justify-center gap-2.5 p-6 text-center">
								<div className="flex size-11 items-center justify-center rounded-xl bg-muted text-muted-foreground">
									<Monitor className="size-5" aria-hidden />
								</div>
								<p className="text-xs font-medium text-muted-foreground">
									{t("downloads_page.desktop_preview_placeholder")}
								</p>
							</div>
						</div>
					</div>

					<div className="space-y-5">
						<p className="text-xs font-semibold uppercase tracking-widest text-muted-foreground">
							{t("downloads_page.eyebrow")}
						</p>
						<h1 className="text-pretty text-4xl font-bold tracking-tight text-foreground sm:text-5xl sm:leading-[1.08]">
							{t("downloads_page.desktop_title")}
						</h1>
						<p className="max-w-lg text-base leading-relaxed text-muted-foreground">
							{t("downloads_page.desktop_description")}
						</p>
						<div className="pt-1">
							<Button
								type="button"
								size="lg"
								className="h-11 gap-2 rounded-full px-6 text-sm font-semibold"
								onClick={handleComingSoon}
							>
								<Download className="size-4" aria-hidden />
								<span>{desktopCtaLabel}</span>
							</Button>
						</div>
					</div>
				</section>

				{/* Row 2: Mobile Showcase */}
				<section className="mt-20 grid items-center gap-10 lg:mt-28 lg:grid-cols-2 lg:gap-20">
					<div className="order-2 space-y-5 lg:order-1">
						<h2 className="text-pretty text-4xl font-bold tracking-tight text-foreground sm:text-5xl sm:leading-[1.08]">
							{t("downloads_page.mobile_title")}
						</h2>
						<p className="max-w-lg text-base leading-relaxed text-muted-foreground">
							{t("downloads_page.mobile_description")}
						</p>
						<div className="flex flex-wrap items-center gap-3 pt-1">
							<button
								type="button"
								onClick={handleComingSoon}
								className="inline-flex h-12 items-center gap-3 rounded-xl bg-zinc-950 px-4 text-left text-white ring-1 ring-white/10 transition-opacity hover:opacity-90 dark:bg-zinc-900"
							>
								<GooglePlayMark />
								<span className="leading-tight">
									<span className="block text-[10px] font-medium uppercase tracking-wider text-zinc-300">
										{t("downloads_page.google_play_kicker")}
									</span>
									<span className="block text-sm font-semibold">Google Play</span>
								</span>
							</button>
							<button
								type="button"
								onClick={handleComingSoon}
								className="inline-flex h-12 items-center gap-3 rounded-xl bg-zinc-950 px-4 text-left text-white ring-1 ring-white/10 transition-opacity hover:opacity-90 dark:bg-zinc-900"
							>
								<AppleMark />
								<span className="leading-tight">
									<span className="block text-[10px] font-medium text-zinc-300">
										{t("downloads_page.app_store_kicker")}
									</span>
									<span className="block text-sm font-semibold">App Store</span>
								</span>
							</button>
						</div>
					</div>

					<div className="order-1 relative aspect-[16/10] w-full overflow-hidden rounded-3xl border border-border/60 bg-gradient-to-br from-sky-400/20 via-cyan-500/15 to-blue-500/20 px-8 pt-8 sm:px-14 sm:pt-10 lg:order-2">
						<div className="mx-auto flex h-full w-full max-w-[260px] flex-col overflow-hidden rounded-t-[2rem] border-x-[6px] border-t-[6px] border-zinc-900/15 bg-card/95 shadow-2xl dark:border-white/15">
							<div className="mx-auto mt-2.5 h-1.5 w-14 rounded-full bg-muted-foreground/25" />
							<div className="flex flex-1 flex-col items-center justify-center gap-2.5 p-6 text-center">
								<div className="flex size-10 items-center justify-center rounded-xl bg-muted text-muted-foreground">
									<Smartphone className="size-5" aria-hidden />
								</div>
								<p className="text-xs font-medium text-muted-foreground">
									{t("downloads_page.mobile_preview_placeholder")}
								</p>
							</div>
						</div>
					</div>
				</section>
			</div>
		</main>
	);
}
