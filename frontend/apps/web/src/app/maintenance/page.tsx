import type { Metadata } from "next";
import { StatusScreen } from "@/components/status/status-screen";

export const metadata: Metadata = {
	title: "503 Maintenance — WiChat",
};

export default function MaintenancePage() {
	return <StatusScreen variant="maintenance" />;
}
