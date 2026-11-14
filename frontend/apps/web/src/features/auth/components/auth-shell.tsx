import Image from "next/image";
import type { ReactNode } from "react";
import { LocaleSwitcher } from "@/features/auth/components/locale-switcher";
import {
	authBrandWordmarkClass,
	authCardInsetXClass,
	authCardShellClass,
	authHeadlineClass,
	authHeadlineInsetXClass,
	authLeadClass,
} from "@/lib/form-field-styles";

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
