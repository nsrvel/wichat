import type { Metadata } from "next";
import Link from "next/link";

export const metadata: Metadata = {
	title: "Privacy Policy — WiChat",
};

export default function PrivacyPage() {
	return (
		<main className="bg-background text-foreground mx-auto max-w-2xl px-6 py-12">
			<h1 className="text-2xl font-semibold tracking-tight">Privacy Policy</h1>
			<p className="mt-6 text-sm leading-relaxed text-muted-foreground">
				This WiChat instance is operated by your organization. Your organization is the data
				controller for personal data processed on this deployment (accounts, messages, files, and
				related metadata).
			</p>
			<p className="mt-4 text-sm leading-relaxed text-muted-foreground">
				WiChat is self-hosted communication software. Policy wording, contact details, retention, and
				subprocessors are defined by the team that runs your instance—not by the WiChat project
				alone.
			</p>
			<p className="mt-4 text-sm leading-relaxed text-muted-foreground">
				Operators can adapt the repository&apos;s{" "}
				<code className="rounded bg-muted px-1 py-0.5 text-xs">PRIVACY_TEMPLATE.md</code> for a
				full policy published on their site.
			</p>
			<p className="mt-8">
				<Link href="/sign-in" className="text-sm font-medium underline underline-offset-4">
					Back to sign in
				</Link>
			</p>
		</main>
	);
}
