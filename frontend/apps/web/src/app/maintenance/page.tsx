import type { Metadata } from "next";
import { StatusScreen } from "@/features/status/status-screen";

export const metadata: Metadata = {
	title: "503 Maintenance — WiChat",
};

export default function MaintenancePage() {
	return <StatusScreen variant="maintenance" />;
}
