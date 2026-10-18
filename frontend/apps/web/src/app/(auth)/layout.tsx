import { appSans } from "@/lib/fonts/app-sans";

export default function AuthLayout({ children }: { children: React.ReactNode }) {
	return (
		<div className={`${appSans.className} flex min-h-0 flex-1 flex-col overflow-auto antialiased`}>
			{children}
		</div>
	);
}
