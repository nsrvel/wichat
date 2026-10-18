"use client";

import { useEffect, useRef, useState } from "react";
import Image from "next/image";
import {
	Hash,
	Bookmark,
	MoreHorizontal,
	Sparkle,
	Info,
	SmilePlus,
	Sparkles,
	AtSign,
	Zap,
	Smile,
	Paperclip,
	Mic,
	Maximize2,
	Menu,
} from "lucide-react";
import type { ReactNode } from "react";

/* ---------- mention-aware text ---------- */
function RichText({ text }: { text: string }) {
	const parts = text.split(/(@[A-Za-z][A-Za-z/ .]*?)(?=[,.]?\s|$)/g);
	return (
		<>
			{parts.map((p, i) =>
				p.startsWith("@") ? (
					<span key={i} className="text-[#4C6AA6] font-medium cursor-pointer hover:underline">
						{p}
					</span>
				) : (
					<span key={i}>{p}</span>
				),
			)}
		</>
	);
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
			className={`h-6 min-w-[30px] px-1.5 rounded-full border flex items-center justify-center gap-1 text-xs transition-colors ${
				on ? "bg-[#F0F4FA] border-[#C9D6EC]" : "bg-[#F5F5F6] border-[#EDEDEE] hover:bg-[#eeeeef]"
			}`}
		>
			<span className="text-[13px] leading-none">{emoji}</span>
			{count !== undefined && <span className="font-medium text-[#3f3f46] text-[11px]">{n}</span>}
		</button>
	);
}

function AddReaction() {
	return (
		<button className="h-6 w-8 rounded-full border border-[#EDEDEE] bg-white flex items-center justify-center text-[#8e8e93] hover:bg-[#f5f5f6] transition-colors">
			<SmilePlus className="h-3.5 w-3.5" strokeWidth={1.8} />
		</button>
	);
}

/* ---------- figma logo ---------- */
function FigmaLogo() {
	return (
		<svg viewBox="0 0 38 57" className="h-[30px] w-[30px] shrink-0" aria-hidden>
			<path d="M19 0h9.5a9.5 9.5 0 0 1 0 19H19V0z" fill="#1ABCFE" transform="translate(0,19)" />
			<path d="M0 28.5A9.5 9.5 0 0 1 9.5 19H19v19H9.5A9.5 9.5 0 0 1 0 28.5z" fill="#A259FF" />
			<path d="M0 9.5A9.5 9.5 0 0 1 9.5 0H19v19H9.5A9.5 9.5 0 0 1 0 9.5z" fill="#F24E1E" />
			<path d="M19 0h9.5a9.5 9.5 0 0 1 0 19H19V0z" fill="#FF7262" />
			<path d="M0 47.5A9.5 9.5 0 0 1 9.5 38H19v9.5a9.5 9.5 0 1 1-19 0z" fill="#0ACF83" />
		</svg>
	);
}

/* ---------- message ---------- */
function Message({
	avatar,
	name,
	time,
	children,
	reactions,
}: {
	avatar: string;
	name: string;
	time?: string;
	children: ReactNode;
	reactions?: ReactNode;
}) {
	return (
		<div className="flex gap-3 px-4 lg:px-8">
			<Image
				src={avatar}
				alt={name}
				width={40}
				height={40}
				className="mt-0.5 h-10 w-10 shrink-0 rounded-full object-cover"
			/>
			<div className="min-w-0 flex-1">
				<div className="flex items-baseline gap-2">
					<span className="text-sm font-semibold text-[#18181b]">{name}</span>
					{time && <span className="text-[11px] text-[#a1a1aa]">{time}</span>}
				</div>
				<div className="mt-1 text-sm leading-[1.45] text-[#3f3f46]">{children}</div>
				{reactions && <div className="mt-2 flex items-center gap-1.5">{reactions}</div>}
			</div>
		</div>
	);
}

/* ---------- mention popup ---------- */
function MentionPopup() {
	return (
		<div className="absolute left-4 lg:left-8 -top-[26px] w-[min(190px,calc(100%-2rem))] bg-white rounded-xl border border-[#EBEBEC] shadow-[0_12px_32px_rgba(0,0,0,0.10)] p-1.5 z-20">
			<div className="px-2 pt-0.5 pb-1 text-xs text-[#a1a1aa]">Members</div>
			<button className="w-full flex items-center gap-2 px-2 h-11 lg:h-[27px] rounded-md bg-[#F0F0F2]">
				<Image
					src="/avatars/demo/diana.jpg"
					alt="Diana Taylor"
					width={18}
					height={18}
					className="h-[18px] w-[18px] rounded-full object-cover"
				/>
				<span className="text-[13px] font-medium text-[#18181b]">Diana Taylor</span>
			</button>
			<button className="w-full flex items-center gap-2 px-2 h-11 lg:h-[27px] rounded-md hover:bg-[#f5f5f6]">
				<Image
					src="/avatars/demo/daniel.jpg"
					alt="Daniel Anderson"
					width={18}
					height={18}
					className="h-[18px] w-[18px] rounded-full object-cover"
				/>
				<span className="text-[13px] font-medium text-[#18181b]">Daniel Anderson</span>
			</button>
		</div>
	);
}

/* ---------- main ---------- */
export function ChannelView({
	onOpenSidebar,
	onOpenMemberPanel,
}: {
	onOpenSidebar?: () => void;
	onOpenMemberPanel?: () => void;
}) {
	const scrollRef = useRef<HTMLDivElement>(null);
	useEffect(() => {
		// start slightly scrolled so the first message is half-cut, like the reference
		scrollRef.current?.scrollTo({ top: 38 });
	}, []);
	return (
		<div className="flex-1 min-w-0 flex flex-col bg-white h-full">
			{/* header */}
			<div className="h-14 lg:h-[46px] shrink-0 border-b border-[#F0F0F1] flex items-center gap-1.5 px-3 lg:px-4">
				<button
					className="lg:hidden h-11 w-11 -ml-1 flex items-center justify-center text-[#3f3f46] rounded-lg hover:bg-[#f4f4f5]"
					onClick={onOpenSidebar}
					aria-label="Open channels"
				>
					<Menu className="h-5 w-5" strokeWidth={1.8} />
				</button>
				<span className="hidden sm:flex h-[22px] w-[22px] rounded-md bg-[#F4F4F5] items-center justify-center shrink-0">
					<Hash className="h-3 w-3 text-[#71717a]" strokeWidth={2.2} />
				</span>
				<div className="flex items-center gap-1.5 min-w-0 text-sm">
					<span className="font-semibold text-[#18181b] whitespace-nowrap">Website</span>
					<span className="text-[#a1a1aa]">/</span>
					<span className="font-semibold text-[#18181b] whitespace-nowrap">v3.0</span>
					<span className="text-[#a1a1aa]">/</span>
					<span className="font-semibold text-[#18181b] truncate">UI-kit design</span>
				</div>
				<button className="hidden sm:flex h-7 w-7 items-center justify-center text-[#8e8e93] hover:bg-[#f4f4f5] rounded-md">
					<Bookmark className="h-3.5 w-3.5" strokeWidth={1.8} />
				</button>
				<div className="ml-auto flex items-center gap-0.5">
					<button className="h-11 w-11 lg:h-8 lg:w-8 flex items-center justify-center text-[#3f3f46] hover:bg-[#f4f4f5] rounded-lg">
						<MoreHorizontal className="h-[18px] w-[18px]" strokeWidth={2} />
					</button>
					<button className="h-11 w-11 lg:h-8 lg:w-8 flex items-center justify-center text-[#3f3f46] hover:bg-[#f4f4f5] rounded-lg">
						<Sparkle className="h-4 w-4" strokeWidth={1.8} />
					</button>
					<button
						className="h-11 w-11 lg:h-8 lg:w-8 flex items-center justify-center rounded-lg hover:bg-[#f4f4f5]"
						onClick={onOpenMemberPanel}
						aria-label="Thread info"
					>
						<span className="h-[22px] w-[22px] rounded-full bg-[#F1F1F2] flex items-center justify-center">
							<Info className="h-3.5 w-3.5 text-[#52525b]" strokeWidth={2} />
						</span>
					</button>
				</div>
			</div>

			{/* messages */}
			<div
				ref={scrollRef}
				className="flex-1 overflow-y-auto thin-scroll pt-5 pb-4 flex flex-col gap-[26px]"
			>
				<Message
					avatar="/avatars/demo/andrew.jpg"
					name="Andrew M."
					time="1w ago"
					reactions={
						<>
							<Reaction emoji="✌️" count={2} />
							<AddReaction />
						</>
					}
				>
					<p>
						Hey team, I wanted to discuss the custom UI-kit we’re developing for the site redesign.
						We need to finalize some components and make key design decisions to ensure consistency
						across the board. Let’s make sure we cover colors, typography, buttons, and any other
						essential UI elements.
					</p>
					<p className="mt-[9px]">
						<RichText text="@UX/UI @Sophia" />
					</p>
				</Message>

				<Message
					avatar="/avatars/demo/diana.jpg"
					name="Diana T."
					time="2d ago"
					reactions={
						<>
							<Reaction emoji="❤️" />
							<AddReaction />
						</>
					}
				>
					<p>
						I have already prepared all styles and components according to our standards during the
						design phase, so the UI kit is 90% complete. All that remains is to add some states to
						the interactive elements and prepare the Lottie files for animations.
					</p>
					<p className="mt-[9px]">
						<RichText text="@Emily D., please take a look and let me know if you have any questions." />
					</p>

					{/* figma link card */}
					<div className="mt-2.5 max-w-[460px] rounded-xl border border-[#ECECED] bg-white shadow-[0_1px_2px_rgba(0,0,0,0.04)] flex items-center gap-2.5 p-2">
						<FigmaLogo />
						<div className="min-w-0 flex-1">
							<div className="text-[13px] font-medium text-[#18181b] truncate">
								Conceptzilla website v.3.0
							</div>
							<div className="text-xs text-[#a1a1aa]">www.figma.com</div>
						</div>
						<button className="shrink-0 h-[30px] px-3.5 rounded-lg border border-[#E4E4E7] bg-white text-[13px] font-medium text-[#18181b] shadow-[0_1px_2px_rgba(0,0,0,0.05)] hover:bg-[#fafafa] transition-colors">
							Quick view
						</button>
					</div>
				</Message>

				<Message avatar="/avatars/demo/daniel.jpg" name="Daniel A." time="3h ago">
					<p>
						<RichText text="Okay, keep me updated. @Diana T. I also wanted to remind you to keep the layers organized." />
					</p>
				</Message>

				<Message
					avatar="/avatars/demo/emily.jpg"
					name="Emily D."
					time="1h ago"
					reactions={
						<>
							<Reaction emoji="🎉" count={3} />
							<AddReaction />
						</>
					}
				>
					<p>
						<RichText text="Great job @Diana T.! Once you finish the states and we'll start development. The handoff looks much smoother than the last time. We're breaking all records 🚀" />
					</p>
				</Message>
			</div>

			{/* composer */}
			<div className="shrink-0 px-3 lg:px-8 pb-4 lg:pb-5 pt-1 relative">
				<MentionPopup />
				<div className="relative rounded-2xl border border-[#EBEBEC] bg-white shadow-[0_2px_10px_rgba(0,0,0,0.05)]">
					<button className="absolute top-2.5 right-2.5 h-7 w-7 flex items-center justify-center text-[#a1a1aa] hover:text-[#52525b]">
						<Maximize2 className="h-3.5 w-3.5" strokeWidth={1.8} />
					</button>
					<div className="px-3.5 pt-3 pb-1.5 text-sm leading-6">
						<p className="text-[#a1a1aa] select-none">
							…finish the states and we’ll start development.
						</p>
						<p className="text-[#a1a1aa] select-none">
							…than the last time. We’re breaking all records
						</p>
						<p className="mt-1 text-[#18181b]">
							💪💪 <span className="text-[#4C6AA6] font-medium">@D</span>
							<span className="caret-blink inline-block w-[2px] h-4 bg-[#18181b] align-[-2px] ml-px" />
						</p>
					</div>
					<div className="flex items-center gap-0 px-2 pb-2">
						{[Sparkles, AtSign, Zap, Smile, Paperclip, Mic].map((Icon, i) => (
							<button
								key={i}
								className="h-11 w-11 lg:h-8 lg:w-8 flex items-center justify-center text-[#71717a] hover:bg-[#f4f4f5] rounded-lg transition-colors"
							>
								<Icon className="h-4 w-4" strokeWidth={1.7} />
							</button>
						))}
						<div className="ml-auto flex items-center gap-1.5">
							<button className="h-[30px] px-3.5 rounded-lg border border-[#E4E4E7] bg-white text-[13px] font-medium text-[#18181b] hover:bg-[#fafafa] transition-colors">
								Discard
							</button>
							<button className="h-[30px] px-4 rounded-lg bg-[#171717] text-white text-[13px] font-medium hover:bg-[#2a2a2a] transition-colors">
								Send
							</button>
						</div>
					</div>
				</div>
			</div>
		</div>
	);
}
