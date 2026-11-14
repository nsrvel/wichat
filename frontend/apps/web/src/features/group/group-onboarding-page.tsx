"use client";

import { useRouter } from "next/navigation";
import Image from "next/image";
import { useTranslation } from "react-i18next";
import { Button } from "@/components/ui/button";
import { authBrandWordmarkClass } from "@/lib/form-field-styles";
import { GroupOnboardingFlow } from "@/features/group/group-onboarding-flow";
import { useLogoutMutation } from "@/services/auth/auth.mutation";

function OnboardingHeader({ onSignOut }: { onSignOut: () => void }) {
	const { t } = useTranslation();
	return (
		<header className="flex shrink-0 items-center justify-between gap-3 border-b border-border bg-background px-4 py-3 sm:px-6 lg:hidden">
			<div className="flex min-w-0 items-center gap-2.5">
				<div className="relative size-8 shrink-0 overflow-hidden rounded-lg border border-border bg-card">
					<Image
						src="/logo_dark.jpg"
						alt=""
						fill
						sizes="2rem"
						className="object-cover dark:hidden"
					/>
					<Image
						src="/logo_light.jpg"
						alt=""
						fill
						sizes="2rem"
						className="hidden object-cover dark:block"
					/>
				</div>
				<span className={authBrandWordmarkClass}>WiChat</span>
			</div>
			<div className="flex shrink-0 items-center gap-0.5">
				<Button
					type="button"
					variant="ghost"
					size="sm"
					className="text-muted-foreground"
					onClick={onSignOut}
				>
					{t("logout")}
				</Button>
			</div>
		</header>
	);
}

/** Full-page onboarding (legacy); shell uses `GroupOnboardingDialog` on `/home`. */
export function GroupOnboardingPage() {
	const router = useRouter();
	const logout = useLogoutMutation();

	const signOut = () => {
		logout.mutate(undefined, { onSettled: () => router.replace("/sign-in") });
	};

	return (
		<div className="relative flex min-h-0 min-w-0 flex-1 flex-col bg-background text-foreground">
			<OnboardingHeader onSignOut={signOut} />
			<div className="overlay-scroll flex min-h-0 flex-1 flex-col items-center justify-center overflow-y-auto px-4 py-8 sm:px-6 sm:py-12">
				<GroupOnboardingFlow presentation="page" />
			</div>
		</div>
	);
}
