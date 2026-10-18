import { ChevronLeft } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import type { ReactNode } from "react";
import { AuthThemeToggle } from "@/features/auth/components/auth-theme-toggle";
import { LocaleSwitcher } from "@/features/auth/components/locale-switcher";

export function AuthShell({
	headline,
	lead,
	notice,
	backHref,
	backLabel,
	children,
}: {
	headline: string;
	lead?: ReactNode;
	notice?: string | null;
	backHref?: string;
	backLabel?: string;
	children: ReactNode;
}) {
	return (
		<main className="relative flex min-h-0 flex-1 flex-col overflow-hidden bg-[#f4f4f4] text-[#171717] dark:bg-[#141414] dark:text-white">
			<header className="relative z-10 flex items-center justify-between px-4 py-4 sm:px-6">
				{backHref && backLabel ? (
					<Link
						href={backHref}
						className="inline-flex items-center gap-1 rounded-lg border border-black/10 bg-white/70 px-2.5 py-1.5 text-sm text-[#3a3a3a] dark:border-white/10 dark:bg-white/5 dark:text-[#d6d6d6]"
					>
						<ChevronLeft className="size-4" />
						{backLabel}
					</Link>
				) : (
					<span />
				)}
				<div className="flex items-center gap-1">
					<AuthThemeToggle />
					<LocaleSwitcher />
				</div>
			</header>
			<div className="relative z-10 mx-auto flex w-full max-w-[400px] flex-1 flex-col items-center justify-center px-4 pb-16">
				<div className="relative size-16 overflow-hidden rounded-2xl">
					<Image
						src="/logo_light.jpg"
						alt="Wirang"
						fill
						priority
						className="object-cover dark:hidden"
					/>
					<Image
						src="/logo_dark.jpg"
						alt=""
						fill
						priority
						aria-hidden
						className="hidden object-cover dark:block"
					/>
				</div>
				<h1 className="mt-6 text-center text-[1.75rem] leading-tight font-semibold tracking-tight">
					{headline}
				</h1>
				{lead ? (
					<div className="mt-2 text-center text-sm text-[#6d6d6d] dark:text-[#a3a3a3]">{lead}</div>
				) : null}
				{notice ? (
					<p className="border-destructive/30 bg-destructive/10 text-destructive mt-4 w-full rounded-lg border px-3 py-2 text-center text-xs">
						{notice}
					</p>
				) : null}
				<div className="mt-8 w-full">{children}</div>
			</div>
		</main>
	);
}

export function AuthOr({ label }: { label: string }) {
	return (
		<div className="my-5 flex items-center gap-3">
			<div className="h-px flex-1 bg-black/10 dark:bg-white/10" />
			<span className="text-xs text-[#8a8a8a]">{label}</span>
			<div className="h-px flex-1 bg-black/10 dark:bg-white/10" />
		</div>
	);
}

export const authInputClass =
	"h-11 rounded-lg border-black/10 bg-white px-3.5 shadow-none placeholder:text-[#9a9a9a] focus-visible:border-black/30 focus-visible:ring-0 dark:border-white/10 dark:bg-[#2a2a2a] dark:text-white dark:placeholder:text-[#8d8d8d] dark:focus-visible:border-white/25";

export const authSubmitClass =
	"h-11 w-full rounded-lg bg-[#171717] text-sm font-medium text-white hover:bg-[#171717]/90 dark:bg-white dark:text-[#171717] dark:hover:bg-white/90";

export const authLinkClass =
	"font-medium text-[#171717] underline underline-offset-4 dark:text-white";
