import Image from "next/image";
import type { ReactNode } from "react";
import { LocaleSwitcher } from "@/features/auth/components/locale-switcher";

/** Login / register card shell — reuse for group onboarding modal. */
export const authCardShellClass =
	"mx-auto w-full min-w-0 max-w-md rounded-[14px] border border-border bg-card py-4 text-left text-card-foreground shadow-elevated sm:rounded-2xl sm:py-7 sm:shadow-elevated-lg md:py-9 max-h-[880px]:py-3.5 max-h-[880px]:sm:py-5 max-h-[880px]:md:py-6";

/** Slightly wider than login card — group create/join modal + full-page onboarding. */
export const groupOnboardingCardShellClass =
	"mx-auto w-full min-w-0 max-w-lg rounded-[14px] border border-border bg-card py-4 text-left text-card-foreground shadow-elevated sm:rounded-2xl sm:py-7 sm:shadow-elevated-lg md:py-9 max-h-[880px]:py-3.5 max-h-[880px]:sm:py-5 max-h-[880px]:md:py-6";

/** Horizontal inset aligned with auth card body (matches form field width). */
export const authCardInsetXClass =
	"px-4 sm:px-7 md:px-9 max-h-[880px]:px-3.5 max-h-[880px]:sm:px-5 max-h-[880px]:md:px-6";

/** Intro copy (headline + lead) — roomier than the form card. */
export const authHeadlineInsetXClass =
	"px-6 sm:px-10 md:px-12 max-h-[880px]:px-5 max-h-[880px]:sm:px-8 max-h-[880px]:md:px-10";

export const authHeadlineClass =
	"w-full min-w-0 max-w-full text-pretty font-normal leading-[1.08] tracking-tight text-foreground [font-family:var(--font-serif,Georgia,serif)] text-[length:clamp(1.375rem,7.5cqi+0.35rem,2.75rem)]";

/** Product / workspace titles on auth surfaces (invite group name, etc.) — sans, not marketing serif. */
export const authProductHeadlineClass =
	"w-full min-w-0 max-w-full text-pretty text-[length:clamp(1.5rem,6.5cqi+0.4rem,2.25rem)] font-semibold leading-tight tracking-tight text-foreground";

export const authLeadClass =
	"mx-auto w-full min-w-0 max-w-[min(100%,34ch)] text-pretty text-xs leading-relaxed font-light text-muted-foreground sm:max-w-[min(100%,38ch)] sm:text-sm md:max-w-[min(100%,42ch)] [&_p]:m-0";

export const authLeadSwitchClass =
	"inline-flex max-w-full flex-wrap items-baseline justify-center gap-[0.1875rem] px-1";

export const authBrandWordmarkClass =
	"truncate text-2xl font-normal tracking-tight text-foreground [font-family:var(--font-serif,Georgia,serif)]";

/** Full-viewport auth background when shell content is not rendered (e.g. session redirect). */
export const authPageBackdropClass =
	"flex min-h-0 min-w-0 flex-1 flex-col bg-background text-foreground";

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
		<main className="overlay-scroll relative flex min-h-0 min-w-0 flex-1 flex-col overflow-x-hidden overflow-y-auto bg-background pt-[max(0.75rem,env(safe-area-inset-top))] pr-[max(1rem,env(safe-area-inset-right))] pb-[max(1.25rem,env(safe-area-inset-bottom))] pl-[max(1rem,env(safe-area-inset-left))] text-foreground sm:px-6 sm:pt-6 sm:pb-8 md:px-10 md:pb-10">
			<header className="relative z-10 flex min-w-0 items-center justify-between gap-3">
				<div className="flex min-w-0 items-center gap-3" aria-label="WiChat" title="WiChat">
					<div className="relative size-9 shrink-0 overflow-hidden rounded-xl border border-border bg-card sm:size-10">
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

						<div className={`${authCardShellClass} ${authCardInsetXClass}`}>
							{notice ? (
								<p className="border-destructive/30 bg-destructive/10 text-destructive mb-4 rounded-xl border p-3 text-xs">
									{notice}
								</p>
							) : null}
							{cardIntro ? <p className="mb-4 text-sm text-muted-foreground">{cardIntro}</p> : null}
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
			<div className="h-px min-w-0 flex-1 bg-border" />
			<span className="shrink-0 text-[10px] tracking-widest text-subtle-foreground uppercase">
				{label}
			</span>
			<div className="h-px flex-1 bg-border" />
		</div>
	);
}

export const authFormClass =
	"flex min-w-0 flex-col gap-2.5 sm:gap-3.5 max-h-[880px]:gap-2 max-h-[880px]:sm:gap-2.5";

/** Sentence-case labels; pass `labelClassName` to override (e.g. uppercase) per field. */
export const authFieldLabelClass =
	"block text-[13px] leading-snug font-medium text-foreground sm:text-sm max-h-[880px]:text-[12px] max-h-[880px]:sm:text-[13px]";

export const authInputClass = [
	"h-10 min-h-10 w-full min-w-0 max-w-full rounded-[10px] border border-transparent px-3 text-sm text-field-foreground shadow-none outline-none transition-[border-color]",
	"sm:h-11 sm:min-h-11 sm:rounded-xl sm:border-2 sm:px-3.5",
	"max-h-[880px]:h-9 max-h-[880px]:min-h-9 max-h-[880px]:sm:h-10 max-h-[880px]:sm:min-h-10",
	"bg-field focus:bg-field focus-visible:bg-field",
	"placeholder:text-placeholder-foreground",
	"focus:border-foreground/20 focus:ring-0 focus-visible:border-foreground/20 focus-visible:ring-0",
	"aria-invalid:border-destructive/50 aria-invalid:focus:border-destructive/60 aria-invalid:focus-visible:border-destructive/60",
].join(" ");

/** Ghost `CustomSelect` trigger (e.g. header locale) — same focus ring habit as inputs, no fill. */
export const authSelectGhostTriggerClass = [
	"h-8 min-h-8 w-auto min-w-0 max-w-none rounded-md border border-transparent bg-transparent px-1.5 shadow-none outline-none",
	"text-[13px] font-semibold tracking-[0.06em] text-muted-foreground uppercase tabular-nums",
	"hover:bg-foreground/5 hover:text-foreground",
	"focus:border-transparent focus:ring-0 focus-visible:border-transparent focus-visible:ring-2 focus-visible:ring-[var(--focus-ring)]",
].join(" ");

/** One OTP cell — full grid cell width, same height/radius/colors as `authInputClass`. */
export const authOtpCellClass = [
	"box-border h-10 min-h-10 w-full min-w-0 rounded-[10px] border border-transparent px-0 text-center text-sm font-medium uppercase tabular-nums leading-none text-field-foreground shadow-none outline-none transition-[border-color]",
	"sm:h-11 sm:min-h-11 sm:rounded-xl sm:border-2",
	"max-h-[880px]:h-9 max-h-[880px]:min-h-9 max-h-[880px]:sm:h-10 max-h-[880px]:sm:min-h-10",
	"bg-field",
	"focus:border-foreground/20 focus:ring-0 focus-visible:border-foreground/20 focus-visible:ring-0",
	"appearance-none [-webkit-appearance:none]",
].join(" ");

/** Shared height/radius for primary and ghost auth actions (login submit, onboarding footer). */
export const authSubmitShapeClass =
	"h-10 min-h-10 rounded-[10px] text-sm font-medium sm:h-11 sm:min-h-11 sm:rounded-xl max-h-[880px]:h-9 max-h-[880px]:min-h-9 max-h-[880px]:sm:h-10 max-h-[880px]:sm:min-h-10";

export const authSubmitClass = [
	authSubmitShapeClass,
	"w-full min-w-0 bg-primary text-primary-foreground hover:bg-primary/90",
].join(" ");

/** Same footprint as `authSubmitClass`; pair with `Button` `variant="ghost"`. */
export const authSubmitGhostClass = [authSubmitShapeClass, "w-full min-w-0"].join(" ");

export const authLinkClass =
	"font-semibold text-primary underline-offset-4 transition-colors hover:underline focus-visible:ring-2 focus-visible:ring-primary/40 focus-visible:outline-none rounded px-0.5 py-0.5";

/** Tertiary links in form rows (e.g. forgot password) — quieter than `authLinkClass`. */
export const authFieldLinkClass =
	"shrink-0 text-[12px] font-normal text-muted-foreground underline-offset-2 transition-colors hover:text-foreground hover:underline focus-visible:ring-2 focus-visible:ring-primary/25 focus-visible:outline-none rounded px-0.5 py-0.5 sm:text-[13px]";
