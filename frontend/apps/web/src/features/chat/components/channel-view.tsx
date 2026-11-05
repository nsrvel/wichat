"use client";

import { createContext, useContext, useEffect, useRef, useState } from "react";
import Image from "next/image";
import { SmilePlus } from "lucide-react";
import type { ReactNode } from "react";
import { ChatComposer } from "@/features/chat/components/chat-composer";
import { ConversationHeader } from "@/features/chat/components/conversation-header";
import { useShellNavigation } from "@/components/layout/shell-navigation";
import { cn } from "@/lib/utils";

const demoAvatar = {
	self: "/avatars/demo/marvell.png",
	yadi: "/avatars/demo/koh-yadi.jpg",
	yanto: "/avatars/demo/koh-yanto.jpg",
} as const;

const demoName = {
	self: "ⱮąɾѵҽӀӀ",
	yadi: "Koh Yadi",
	yanto: "Koh Yanto",
} as const;

/* ---------- mention-aware text ---------- */
function escapeRegExp(value: string) {
	return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function buildMentionPattern(handles: readonly string[]) {
	const labels = [...handles].sort((a, b) => b.length - a.length).map(escapeRegExp);
	return new RegExp(`@(?:${labels.join("|")})`, "g");
}

const demoMentionPattern = buildMentionPattern([demoName.yadi, demoName.yanto, demoName.self]);

type MentionTone = "incoming" | "outgoing";

function RichText({ text, tone = "incoming" }: { text: string; tone?: MentionTone }) {
	const mentionClass =
		tone === "outgoing"
			? "font-semibold text-chat-mention-outgoing"
			: "font-semibold text-chat-mention";

	const nodes: ReactNode[] = [];
	let lastIndex = 0;
	for (const match of text.matchAll(demoMentionPattern)) {
		const index = match.index ?? 0;
		if (index > lastIndex) {
			nodes.push(<span key={`t-${lastIndex}`}>{text.slice(lastIndex, index)}</span>);
		}
		const mention = match[0];
		nodes.push(
			<span key={`m-${index}`} className={cn(mentionClass, "cursor-pointer hover:underline")}>
				{mention}
			</span>,
		);
		lastIndex = index + mention.length;
	}
	if (lastIndex < text.length) {
		nodes.push(<span key={`t-${lastIndex}`}>{text.slice(lastIndex)}</span>);
	}
	if (nodes.length === 0) {
		return <>{text}</>;
	}
	return <>{nodes}</>;
}

const MentionToneContext = createContext<MentionTone>("incoming");

function MentionToneProvider({ tone, children }: { tone: MentionTone; children: ReactNode }) {
	return <MentionToneContext.Provider value={tone}>{children}</MentionToneContext.Provider>;
}

function MessageRichText({ text }: { text: string }) {
	const tone = useContext(MentionToneContext);
	return <RichText text={text} tone={tone} />;
}

/* ---------- reactions ---------- */
function Reaction({ emoji, count, active }: { emoji: string; count?: number; active?: boolean }) {
	const [n, setN] = useState(count ?? 0);
	const [on, setOn] = useState(!!active);
	return (
		<button
			onClick={() => {
				setN(on ? n - 1 : n + 1);
				setOn(!on);
			}}
			className={cn(
				"flex h-6 min-w-[30px] items-center justify-center gap-1 rounded-full border px-1.5 text-xs transition-colors",
				on
					? "border-[#c9d6ec] bg-[#f0f4fa] dark:border-[#3d4f6f] dark:bg-[#1e2a3d]"
					: "border-chat-surface-border bg-chat-composer-field hover:bg-sidebar-accent",
			)}
		>
			<span className="text-[13px] leading-none">{emoji}</span>
			{count !== undefined ? (
				<span className="text-[11px] font-medium text-muted-foreground">{n}</span>
			) : null}
		</button>
	);
}

function AddReaction() {
	return (
		<button
			type="button"
			className="flex h-6 w-8 items-center justify-center rounded-full border border-chat-surface-border bg-chat-surface text-muted-foreground transition-colors hover:bg-sidebar-accent"
		>
			<SmilePlus className="h-3.5 w-3.5" strokeWidth={1.8} />
		</button>
	);
}

/* ---------- message (bubble) ---------- */
function Message({
	avatar,
	name,
	time,
	children,
	reactions,
	mine = false,
}: {
	avatar: string;
	name: string;
	time?: string;
	children: ReactNode;
	reactions?: ReactNode;
	mine?: boolean;
}) {
	const mentionTone: MentionTone = mine ? "outgoing" : "incoming";
	const bubbleClass = cn(
		"rounded-2xl border px-3.5 py-2.5 text-sm leading-[1.45] text-foreground [&_p+p]:mt-2",
		mine
			? "rounded-tr-md border-chat-bubble-outgoing-border bg-chat-bubble-outgoing"
			: "rounded-tl-md border-chat-bubble-border bg-chat-bubble",
	);

	if (mine) {
		return (
			<div className="flex justify-end">
				<div className="min-w-0 max-w-[min(88%,34rem)]">
					{time ? (
						<div className="mb-1 flex justify-end px-0.5">
							<span className="text-[11px] text-muted-foreground">{time}</span>
						</div>
					) : null}
					<div className={bubbleClass}>
						<MentionToneProvider tone={mentionTone}>{children}</MentionToneProvider>
					</div>
					{reactions ? (
						<div className="mt-1.5 flex flex-wrap justify-end gap-1.5 px-0.5">{reactions}</div>
					) : null}
				</div>
			</div>
		);
	}

	return (
		<div className="flex gap-2.5">
			<Image
				src={avatar}
				alt={name}
				width={32}
				height={32}
				className="mt-0.5 h-8 w-8 shrink-0 rounded-full object-cover ring-1 ring-chat-surface-border"
			/>
			<div className="min-w-0 max-w-full flex-1">
				<div className="mb-1 flex items-baseline gap-2 px-0.5">
					<span className="text-[13px] font-semibold text-foreground">{name}</span>
					{time ? <span className="text-[11px] text-muted-foreground">{time}</span> : null}
				</div>
				<div className={bubbleClass}>
					<MentionToneProvider tone={mentionTone}>{children}</MentionToneProvider>
				</div>
				{reactions ? (
					<div className="mt-1.5 flex flex-wrap items-center gap-1.5 px-0.5">{reactions}</div>
				) : null}
			</div>
		</div>
	);
}

/* ---------- main ---------- */
export function ChannelView({ onOpenSidebar }: { onOpenSidebar?: () => void }) {
	const { active } = useShellNavigation();
	const scrollRef = useRef<HTMLDivElement>(null);
	useEffect(() => {
		// start slightly scrolled so the first message is half-cut, like the reference
		scrollRef.current?.scrollTo({ top: 38 });
	}, []);
	return (
		<div className="flex h-full min-w-0 flex-1 flex-col bg-chat-surface text-foreground">
			<div className="lg:hidden">
				<ConversationHeader active={active} onOpenSidebar={onOpenSidebar} />
			</div>

			<div className="flex min-h-0 flex-1 flex-col">
				<div ref={scrollRef} className="flex-1 overflow-y-auto thin-scroll">
					<div className="mx-auto flex w-full max-w-[56rem] flex-col gap-5 px-4 py-4 sm:px-5 lg:max-w-[60rem] lg:px-6">
						<Message
							avatar={demoAvatar.yadi}
							name={demoName.yadi}
							time="1 minggu lalu"
							reactions={
								<>
									<Reaction emoji="👍" count={2} />
									<AddReaction />
								</>
							}
						>
							<p>
								Tim, minggu ini kita fokus polish shell chat WiChat dulu ya. Sidebar, header
								channel, sama composer harus konsisten warnanya sebelum nyambung ke backend beneran.
							</p>
							<p className="mt-[9px]">
								<MessageRichText text="@Koh Yanto @ⱮąɾѵҽӀӀ" />
							</p>
						</Message>

						<Message
							avatar={demoAvatar.yanto}
							name={demoName.yanto}
							time="2 hari lalu"
							reactions={
								<>
									<Reaction emoji="🔥" />
									<AddReaction />
								</>
							}
						>
							<p>
								Siap. Icon channel WiChat sudah pakai tile bulat yang sama kayak di sidebar — jadi
								header nggak pakai logo workspace lagi. Lebih enak dibaca pas pindah-pindah channel.
							</p>
						</Message>

						<Message avatar={demoAvatar.self} name={demoName.self} time="2 hari lalu" mine>
							<p>
								Mantap. Nanti tab Chat / Pinned / Shared juga jangan sampai geser alignment-nya pas
								resize sidebar. Kemarin sempat aneh pas nav stretch.
							</p>
						</Message>

						<Message
							avatar={demoAvatar.yadi}
							name={demoName.yadi}
							time="1 hari lalu"
							reactions={
								<>
									<Reaction emoji="❤️" count={1} />
									<AddReaction />
								</>
							}
						>
							<p>
								Compose bar-nya udah mirip Slack — satu baris, warna bubble. Tinggal wiring kirim
								pesan ke ms-chat kalau gateway-nya ready.
							</p>
							<p className="mt-[9px]">
								<MessageRichText text="@ⱮąɾѵҽӀӀ tolong cek juga dock user bawah: background-nya harus sama chat surface." />
							</p>
						</Message>

						<Message avatar={demoAvatar.yanto} name={demoName.yanto} time="5 jam lalu">
							<p>
								<MessageRichText text="Noted. @Koh Yadi aku lagi siapin fixture DM kita berdua biar demo-nya hidup." />
							</p>
						</Message>

						<Message
							avatar={demoAvatar.self}
							name={demoName.self}
							time="3 jam lalu"
							mine
							reactions={
								<>
									<Reaction emoji="🎉" count={3} />
									<AddReaction />
								</>
							}
						>
							<p>
								Oke semua. Isi channel WiChat sekarang pakai kita bertiga aja biar realistis —
								avatar ngikut sidebar. Next slice: websocket + history scroll.
							</p>
						</Message>

						<Message avatar={demoAvatar.yadi} name={demoName.yadi} time="1 jam lalu">
							<p>
								Sip, lanjut pelan-pelan. Kalau ada yang blocking di port lokal kabarin di sini aja,
								jangan di thread lain biar ke-track.
							</p>
						</Message>
					</div>
				</div>

				<ChatComposer />
			</div>
		</div>
	);
}
