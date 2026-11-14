import type { Metadata } from "next";
import { StatusScreen } from "@/components/status/status-screen";

export const metadata: Metadata = {
	title: "403 Forbidden — WiChat",
};

export default function ForbiddenPage() {
	return <StatusScreen variant="forbidden" />;
}
