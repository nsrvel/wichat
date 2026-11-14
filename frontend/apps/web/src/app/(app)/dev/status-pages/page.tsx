import type { Metadata } from "next";
import { StatusPagesPreview } from "@/features/dev/status-pages/status-pages-preview";

export const metadata: Metadata = {
	title: "Status Pages Preview — WiChat",
	description: "Interactive preview for WiChat 404, 403, 401, 500, and 503 status pages.",
};

export default function DevStatusPagesPage() {
	return <StatusPagesPreview />;
}
