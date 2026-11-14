"use client";

import { useEffect, useRef } from "react";

type TurnstileAPI = {
	render: (
		container: HTMLElement,
		options: {
			sitekey: string;
			theme?: "auto" | "light" | "dark";
			callback?: (token: string) => void;
			"expired-callback"?: () => void;
			"error-callback"?: () => void;
		},
	) => string;
	reset: (widgetId?: string) => void;
	remove: (widgetId?: string) => void;
};

declare global {
	interface Window {
		turnstile?: TurnstileAPI;
	}
}

const TURNSTILE_SCRIPT_ID = "wichat-turnstile-script";
const TURNSTILE_SCRIPT_SRC =
	"https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit";

type TurnstileWidgetProps = {
	siteKey: string;
	resetKey?: number;
	onToken: (token: string) => void;
};

export function TurnstileWidget({ siteKey, resetKey = 0, onToken }: TurnstileWidgetProps) {
	const containerRef = useRef<HTMLDivElement>(null);
	const widgetIdRef = useRef<string | null>(null);
	const onTokenRef = useRef(onToken);

	useEffect(() => {
		onTokenRef.current = onToken;
	}, [onToken]);

	useEffect(() => {
		if (!siteKey) {
			return;
		}
		let cancelled = false;

		const mountWidget = () => {
			if (cancelled || !containerRef.current || !window.turnstile || widgetIdRef.current) {
				return;
			}
			widgetIdRef.current = window.turnstile.render(containerRef.current, {
				sitekey: siteKey,
				theme: "auto",
				callback: (token) => onTokenRef.current(token),
				"expired-callback": () => onTokenRef.current(""),
				"error-callback": () => onTokenRef.current(""),
			});
		};

		if (window.turnstile) {
			mountWidget();
		} else {
			let script = document.getElementById(TURNSTILE_SCRIPT_ID) as HTMLScriptElement | null;
			if (!script) {
				script = document.createElement("script");
				script.id = TURNSTILE_SCRIPT_ID;
				script.src = TURNSTILE_SCRIPT_SRC;
				script.async = true;
				script.defer = true;
				document.head.appendChild(script);
			}
			script.addEventListener("load", mountWidget);
			return () => {
				cancelled = true;
				script?.removeEventListener("load", mountWidget);
				if (widgetIdRef.current) {
					window.turnstile?.remove(widgetIdRef.current);
					widgetIdRef.current = null;
				}
			};
		}

		return () => {
			cancelled = true;
			if (widgetIdRef.current) {
				window.turnstile?.remove(widgetIdRef.current);
				widgetIdRef.current = null;
			}
		};
	}, [siteKey]);

	useEffect(() => {
		if (resetKey > 0 && widgetIdRef.current) {
			window.turnstile?.reset(widgetIdRef.current);
		}
	}, [resetKey]);

	if (!siteKey) {
		return null;
	}

	return <div ref={containerRef} className="flex justify-center" />;
}
