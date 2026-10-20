import Image from "next/image";
import type { ReactNode } from "react";
import { AuthThemeToggle } from "@/features/auth/components/auth-theme-toggle";
import { LocaleSwitcher } from "@/features/auth/components/locale-switcher";
/** Horizontal inset aligned with auth card body (matches form field width). */
export const authCardInsetXClass =
	"px-4 sm:px-7 md:px-9 max-h-[880px]:px-3.5 max-h-[880px]:sm:px-5 max-h-[880px]:md:px-6";

/** Intro copy (headline + lead) — roomier than the form card. */
export const authHeadlineInsetXClass =
	"px-6 sm:px-10 md:px-12 max-h-[880px]:px-5 max-h-[880px]:sm:px-8 max-h-[880px]:md:px-10";

export const authHeadlineClass =
	"w-full min-w-0 max-w-full text-pretty font-normal leading-[1.08] tracking-tight [font-family:var(--font-serif,Georgia,serif)] text-[length:clamp(1.375rem,7.5cqi+0.35rem,2.75rem)]";

export const authLeadClass =
	"mx-auto w-full min-w-0 max-w-[min(100%,34ch)] text-pretty text-xs leading-relaxed font-light text-[#6d6d6d] sm:max-w-[min(100%,38ch)] sm:text-sm md:max-w-[min(100%,42ch)] dark:text-[#a3a3a3] [&_p]:m-0";

export const authLeadSwitchClass =
	"inline-flex max-w-full flex-wrap items-baseline justify-center gap-[0.1875rem] px-1";

export const authBrandWordmarkClass =
	"truncate text-2xl font-normal tracking-tight [font-family:var(--font-serif,Georgia,serif)] text-[#171717] dark:text-white";

export function AuthShell({
	headline,
	lead,
	inviteBanner,
	notice,
	cardIntro,
	children,
}: {
	headline: ReactNode;
	lead?: ReactNode;
	inviteBanner?: ReactNode;
	notice?: string | null;
	cardIntro?: ReactNode;
	children: ReactNode;
}) {
	return (
		<main className="relative flex min-h-0 min-w-0 flex-1 flex-col overflow-x-hidden overflow-y-auto bg-[#f4f4f4] pt-[max(0.75rem,env(safe-area-inset-top))] pr-[max(1rem,env(safe-area-inset-right))] pb-[max(1.25rem,env(safe-area-inset-bottom))] pl-[max(1rem,env(safe-area-inset-left))] text-[#171717] sm:px-6 sm:pt-6 sm:pb-8 md:px-10 md:pb-10 dark:bg-[#141414] dark:text-white">
			<header className="relative z-10 flex min-w-0 items-center justify-between gap-3">
				<div className="flex min-w-0 items-center gap-3" aria-label="WiChat" title="WiChat">
					<div className="relative size-9 shrink-0 overflow-hidden rounded-xl border border-black/10 bg-white sm:size-10 dark:border-white/10 dark:bg-[#1e1e1e]">
						<Image
							src="/logo_dark.jpg"
							alt=""
							fill
							priority
							sizes="(min-width: 640px) 2.5rem, 2.25rem"
							className="object-cover dark:hidden"
						/>
						<Image
							src="/logo_light.jpg"
							alt=""
							fill
							priority
							sizes="(min-width: 640px) 2.5rem, 2.25rem"
							aria-hidden
							className="hidden object-cover dark:block"
						/>
					</div>
					<span className={authBrandWordmarkClass}>WiChat</span>
				</div>
				<div className="flex shrink-0 items-center gap-0.5 sm:gap-1">
					<AuthThemeToggle />
					<LocaleSwitcher />
				</div>
			</header>

			<div className="relative z-10 flex min-h-0 w-full flex-1 flex-col">
				<div className="mx-auto my-auto flex w-full min-w-0 max-w-xl flex-col items-center px-4 py-4 pb-8 sm:px-6 sm:py-6 sm:pb-10 max-h-[880px]:py-3 max-h-[880px]:pb-6">
					<div className="flex w-full min-w-0 flex-col items-center text-center">
						<div
							className={`@container mb-5 w-full min-w-0 max-w-xl sm:mb-10 max-h-[880px]:mb-4 max-h-[880px]:sm:mb-6 ${authHeadlineInsetXClass}`}
						>
							{inviteBanner}
							<h1
								className={`${authHeadlineClass} mb-2 sm:mb-3 max-h-[880px]:mb-1.5 max-h-[880px]:sm:mb-2`}
							>
								{headline}
							</h1>
							{lead ? <div className={authLeadClass}>{lead}</div> : null}
						</div>

						<div
							className={`mx-auto w-full min-w-0 max-w-md rounded-[14px] border border-black/8 bg-white py-4 text-left shadow-[0_4px_24px_rgba(0,0,0,0.05)] sm:rounded-2xl sm:py-7 sm:shadow-[0_8px_30px_rgba(0,0,0,0.04)] md:py-9 max-h-[880px]:py-3.5 max-h-[880px]:sm:py-5 max-h-[880px]:md:py-6 dark:border-white/8 dark:bg-[#1a1a1a] dark:shadow-[0_8px_30px_rgba(0,0,0,0.3)] ${authCardInsetXClass}`}
						>
							{notice ? (
								<p className="border-destructive/30 bg-destructive/10 text-destructive mb-4 rounded-xl border p-3 text-xs">
									{notice}
								</p>
							) : null}
							{cardIntro ? (
								<p className="mb-4 text-sm text-[#6d6d6d] dark:text-[#a3a3a3]">{cardIntro}</p>
							) : null}
							<div className="space-y-3 sm:space-y-4 max-h-[880px]:space-y-2 max-h-[880px]:sm:space-y-2.5">
								{children}
							</div>
						</div>
					</div>
				</div>
			</div>
		</main>
	);
}

export function AuthOr({ label }: { label: string }) {
	return (
		<div className="flex min-w-0 items-center gap-2 sm:gap-3">
			<div className="h-px min-w-0 flex-1 bg-black/10 dark:bg-white/10" />
			<span className="shrink-0 text-[10px] tracking-widest text-[#8a8a8a] uppercase">{label}</span>
			<div className="h-px flex-1 bg-black/10 dark:bg-white/10" />
		</div>
	);
}

export const authFormClass =
	"flex min-w-0 flex-col gap-2.5 sm:gap-3.5 max-h-[880px]:gap-2 max-h-[880px]:sm:gap-2.5";

/** Sentence-case labels; pass `labelClassName` to override (e.g. uppercase) per field. */
export const authFieldLabelClass =
	"block text-[13px] leading-snug font-medium text-[#171717] sm:text-sm max-h-[880px]:text-[12px] max-h-[880px]:sm:text-[13px] dark:text-[#ececec]";

export const authInputClass = [
	"h-10 min-h-10 w-full min-w-0 max-w-full rounded-[10px] border border-transparent px-3 text-sm text-[#171717] shadow-none outline-none transition-[border-color]",
	"sm:h-11 sm:min-h-11 sm:rounded-xl sm:border-2 sm:px-3.5",
	"max-h-[880px]:h-9 max-h-[880px]:min-h-9 max-h-[880px]:sm:h-10 max-h-[880px]:sm:min-h-10",
	"bg-[#f0f0f0] dark:bg-white/[0.07] focus:bg-[#f0f0f0] focus-visible:bg-[#f0f0f0] dark:focus:bg-white/[0.07] dark:focus-visible:bg-white/[0.07]",
	"placeholder:text-[#9a9a9a] dark:text-white dark:placeholder:text-[#8d8d8d]",
	"focus:border-black/20 focus:ring-0 focus-visible:border-black/20 focus-visible:ring-0",
	"dark:focus:border-white/22 dark:focus-visible:border-white/22",
	"aria-invalid:border-destructive/50 aria-invalid:focus:border-destructive/60 aria-invalid:focus-visible:border-destructive/60 dark:aria-invalid:border-destructive/55",
].join(" ");

/** One OTP cell — full grid cell width, same height/radius/colors as `authInputClass`. */
export const authOtpCellClass = [
	"box-border h-10 min-h-10 w-full min-w-0 rounded-[10px] border border-transparent px-0 text-center text-sm font-medium uppercase tabular-nums leading-none shadow-none outline-none transition-[border-color]",
	"sm:h-11 sm:min-h-11 sm:rounded-xl sm:border-2",
	"max-h-[880px]:h-9 max-h-[880px]:min-h-9 max-h-[880px]:sm:h-10 max-h-[880px]:sm:min-h-10",
	"bg-[#f0f0f0] dark:bg-white/[0.07]",
	"text-[#171717] dark:text-white",
	"focus:border-black/20 focus:ring-0 focus-visible:border-black/20 focus-visible:ring-0",
	"dark:focus:border-white/22 dark:focus-visible:border-white/22",
	"appearance-none [-webkit-appearance:none]",
].join(" ");

export const authSubmitClass =
	"h-10 min-h-10 w-full min-w-0 rounded-[10px] bg-[#171717] text-sm font-medium text-white hover:bg-[#171717]/90 sm:h-11 sm:min-h-11 sm:rounded-xl max-h-[880px]:h-9 max-h-[880px]:min-h-9 max-h-[880px]:sm:h-10 max-h-[880px]:sm:min-h-10 dark:bg-white dark:text-[#171717] dark:hover:bg-white/90";

export const authLinkClass =
	"font-semibold text-primary underline-offset-4 transition-colors hover:underline focus-visible:ring-2 focus-visible:ring-primary/40 focus-visible:outline-none rounded px-0.5 py-0.5";

/** Tertiary links in form rows (e.g. forgot password) — quieter than `authLinkClass`. */
export const authFieldLinkClass =
	"shrink-0 text-[12px] font-normal text-[#6d6d6d] underline-offset-2 transition-colors hover:text-[#171717] hover:underline focus-visible:ring-2 focus-visible:ring-primary/25 focus-visible:outline-none rounded px-0.5 py-0.5 sm:text-[13px] dark:text-[#a3a3a3] dark:hover:text-[#ececec]";
