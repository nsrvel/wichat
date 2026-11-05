export default function AuthLayout({ children }: { children: React.ReactNode }) {
	return (
		<div
			className="flex min-h-0 flex-1 flex-col overflow-auto bg-background text-foreground antialiased"
			suppressHydrationWarning
		>
			{children}
		</div>
	);
}
