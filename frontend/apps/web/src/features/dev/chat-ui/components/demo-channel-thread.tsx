"use client";

import {
	buildMentionPattern,
	MentionToneProvider,
	MessageMentionText,
} from "@/features/dev/chat-ui/lib/mention-text";
import {
	demoAvatar,
	demoMentionHandles,
	demoName,
} from "@/features/dev/chat-ui/lib/demo-channel-messages";
import {
	DemoAddReaction,
	DemoReaction,
} from "@/features/dev/chat-ui/components/demo-message-reactions";
import { MessageRow } from "@/features/chat/components/message-row";

const demoMentionPattern = buildMentionPattern(demoMentionHandles);

export function DemoChannelThread() {
	return (
		<>
			<MessageRow
				authorAvatarUrl={demoAvatar.yadi}
				authorDisplayName={demoName.yadi}
				timeLabel="1 minggu lalu"
				reactions={
					<>
						<DemoReaction emoji="👍" count={2} />
						<DemoAddReaction />
					</>
				}
			>
				<MentionToneProvider tone="incoming">
					<p>
						Tim, minggu ini kita fokus polish shell chat WiChat dulu ya. Sidebar, header channel,
						sama composer harus konsisten warnanya sebelum nyambung ke backend beneran.
					</p>
					<p className="mt-[9px]">
						<MessageMentionText text="@Koh Yanto @ⱮąɾѵҽӀӀ" pattern={demoMentionPattern} />
					</p>
				</MentionToneProvider>
			</MessageRow>

			<MessageRow
				authorAvatarUrl={demoAvatar.yanto}
				authorDisplayName={demoName.yanto}
				timeLabel="2 hari lalu"
				reactions={
					<>
						<DemoReaction emoji="🔥" />
						<DemoAddReaction />
					</>
				}
			>
				<p>
					Siap. Icon channel WiChat sudah pakai tile bulat yang sama kayak di sidebar — jadi header
					nggak pakai logo workspace lagi. Lebih enak dibaca pas pindah-pindah channel.
				</p>
			</MessageRow>

			<MessageRow
				authorDisplayName={demoName.self}
				authorAvatarUrl={demoAvatar.self}
				timeLabel="2 hari lalu"
				mine
			>
				<p>
					Mantap. Nanti tab Chat / Pinned / Shared juga jangan sampai geser alignment-nya pas resize
					sidebar. Kemarin sempat aneh pas nav stretch.
				</p>
			</MessageRow>

			<MessageRow
				authorAvatarUrl={demoAvatar.yadi}
				authorDisplayName={demoName.yadi}
				timeLabel="1 hari lalu"
				reactions={
					<>
						<DemoReaction emoji="❤️" count={1} />
						<DemoAddReaction />
					</>
				}
			>
				<MentionToneProvider tone="incoming">
					<p>
						Compose bar-nya udah mirip Slack — satu baris, warna bubble. Tinggal wiring kirim pesan
						ke ms-chat kalau gateway-nya ready.
					</p>
					<p className="mt-[9px]">
						<MessageMentionText
							text="@ⱮąɾѵҽӀӀ tolong cek juga dock user bawah: background-nya harus sama chat surface."
							pattern={demoMentionPattern}
						/>
					</p>
				</MentionToneProvider>
			</MessageRow>

			<MessageRow
				authorAvatarUrl={demoAvatar.yanto}
				authorDisplayName={demoName.yanto}
				timeLabel="5 jam lalu"
			>
				<MentionToneProvider tone="incoming">
					<p>
						<MessageMentionText
							text="Noted. @Koh Yadi aku lagi siapin fixture DM kita berdua biar demo-nya hidup."
							pattern={demoMentionPattern}
						/>
					</p>
				</MentionToneProvider>
			</MessageRow>

			<MessageRow
				authorAvatarUrl={demoAvatar.self}
				authorDisplayName={demoName.self}
				timeLabel="3 jam lalu"
				mine
				reactions={
					<>
						<DemoReaction emoji="🎉" count={3} />
						<DemoAddReaction />
					</>
				}
			>
				<p>
					Oke semua. Isi channel WiChat sekarang pakai kita bertiga aja biar realistis — avatar
					ngikut sidebar. Next slice: websocket + history scroll.
				</p>
			</MessageRow>

			<MessageRow
				authorAvatarUrl={demoAvatar.yadi}
				authorDisplayName={demoName.yadi}
				timeLabel="1 jam lalu"
			>
				<p>
					Sip, lanjut pelan-pelan. Kalau ada yang blocking di port lokal kabarin di sini aja, jangan
					di thread lain biar ke-track.
				</p>
			</MessageRow>
		</>
	);
}
