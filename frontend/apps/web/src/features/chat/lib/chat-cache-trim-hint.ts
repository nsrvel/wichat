"use client";

import {
	createContext,
	createElement,
	useContext,
	useMemo,
	type MutableRefObject,
	type ReactNode,
} from "react";
import type { CacheTrimHint } from "@/features/chat/lib/message-cache-window";

const ChatCacheTrimHintContext = createContext<MutableRefObject<CacheTrimHint> | null>(null);

export function ChatCacheTrimHintProvider({ children }: { children: ReactNode }) {
	const hintRef = useMemo<MutableRefObject<CacheTrimHint>>(() => ({ current: "near-bottom" }), []);
	return createElement(ChatCacheTrimHintContext.Provider, { value: hintRef }, children);
}

const fallbackHintRef: MutableRefObject<CacheTrimHint> = { current: "near-bottom" };

export function useChatCacheTrimHintRef(): MutableRefObject<CacheTrimHint> {
	const ctx = useContext(ChatCacheTrimHintContext);
	return ctx ?? fallbackHintRef;
}
