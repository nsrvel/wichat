import type { Metadata } from "next";
import { StatusScreen } from "@/components/status/status-screen";

export const metadata: Metadata = {
	title: "404 — WiChat",
};

export default function NotFoundPage() {
	return <StatusScreen variant="not-found" />;
}
