import { env } from "@/lib/env";

export type GatewayConnectionStatus = "connected" | "reconnecting";

type Listener = () => void;

let status: GatewayConnectionStatus = "connected";
const statusListeners = new Set<Listener>();
const recoveredListeners = new Set<Listener>();

let probeTimer: ReturnType<typeof setTimeout> | null = null;
let probeInFlight = false;
let attempt = 0;
let windowListenersAttached = false;

function setStatus(next: GatewayConnectionStatus) {
	if (status === next) {
		return;
	}
	status = next;
	for (const listener of statusListeners) {
		listener();
	}
}

function clearProbeTimer() {
	if (probeTimer !== null) {
		clearTimeout(probeTimer);
		probeTimer = null;
	}
}

function handleWindowWake() {
	if (status === "reconnecting") {
		void probeGatewayNow();
	}
}

function attachWindowListeners() {
	if (windowListenersAttached || typeof window === "undefined") {
		return;
	}
	windowListenersAttached = true;
	window.addEventListener("online", handleWindowWake);
	window.addEventListener("focus", handleWindowWake);
}

function detachWindowListeners() {
	if (!windowListenersAttached || typeof window === "undefined") {
		return;
	}
	windowListenersAttached = false;
	window.removeEventListener("online", handleWindowWake);
	window.removeEventListener("focus", handleWindowWake);
}

function scheduleNextProbe() {
	if (status !== "reconnecting" || probeTimer !== null || typeof window === "undefined") {
		return;
	}
	const delay = Math.min(2000 * Math.pow(1.5, attempt), 5000);
	probeTimer = setTimeout(() => {
		probeTimer = null;
		void probeGatewayNow();
	}, delay);
}

export async function probeGatewayNow(): Promise<boolean> {
	if (typeof window === "undefined" || probeInFlight) {
		return false;
	}
	probeInFlight = true;
	try {
		const res = await fetch(`${env.NEXT_PUBLIC_API_URL}/api/v1/healthz`, {
			method: "GET",
			cache: "no-store",
			signal: AbortSignal.timeout(3000),
		});
		if (res.ok) {
			notifyGatewayReachable();
			return true;
		}
	} catch {
		// Gateway still unreachable
	} finally {
		probeInFlight = false;
	}
	attempt += 1;
	for (const listener of statusListeners) {
		listener();
	}
	scheduleNextProbe();
	return false;
}

export function notifyGatewayUnreachable(): void {
	if (typeof window === "undefined") {
		return;
	}
	if (status !== "reconnecting") {
		attempt = 0;
		setStatus("reconnecting");
		attachWindowListeners();
	}
	scheduleNextProbe();
}

export function notifyGatewayReachable(): void {
	if (status !== "reconnecting") {
		return;
	}
	clearProbeTimer();
	detachWindowListeners();
	attempt = 0;
	setStatus("connected");
	for (const listener of recoveredListeners) {
		listener();
	}
}

export function getGatewayStatus(): GatewayConnectionStatus {
	return status;
}

export function getGatewayAttempt(): number {
	return attempt;
}

export function subscribeGatewayStatus(listener: Listener): () => void {
	statusListeners.add(listener);
	return () => {
		statusListeners.delete(listener);
	};
}

export function onGatewayRecovered(listener: Listener): () => void {
	recoveredListeners.add(listener);
	return () => {
		recoveredListeners.delete(listener);
	};
}
