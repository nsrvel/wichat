import Image from "next/image";
import { LocaleSwitcher } from "@/features/auth/components/locale-switcher";
import { authBrandWordmarkClass } from "@/lib/form-field-styles";

export function BrandHeader() {
	return (
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
	);
}
