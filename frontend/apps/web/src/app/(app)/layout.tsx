import { AppShell } from "@/components/layout/app-shell";
import { AuthGate } from "@/features/auth/auth-gate";

export default function AppLayout({ children }: { children: React.ReactNode }) {
	return (
		<div
			className="flex min-h-0 flex-1 flex-col overflow-hidden antialiased"
			suppressHydrationWarning
		>
			<AuthGate>
				<AppShell>{children}</AppShell>
			</AuthGate>
		</div>
	);
}
