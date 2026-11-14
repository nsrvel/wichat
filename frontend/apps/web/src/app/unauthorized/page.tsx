import type { Metadata } from "next";
import { StatusScreen } from "@/components/status/status-screen";

export const metadata: Metadata = {
	title: "401 Unauthorized — WiChat",
};

export default function UnauthorizedPage() {
	return <StatusScreen variant="unauthorized" />;
}
