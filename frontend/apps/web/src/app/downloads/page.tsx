import type { Metadata } from "next";
import { DownloadsPageView } from "@/features/downloads/downloads-page-view";

export const metadata: Metadata = {
	title: "Downloads — WiChat",
	description: "Download WiChat for Desktop and Mobile.",
};

export default function DownloadsPage() {
	return <DownloadsPageView />;
}
