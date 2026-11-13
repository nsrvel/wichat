# WiChat Role Permissions (Curated & Roadmap)

Dokumen ini adalah hasil saringan dari [discord-permissions-reference.md](discord-permissions-reference.md) yang disesuaikan dengan arsitektur, terminologi (**Group** alih-alih _Server_), dan roadmap fitur **WiChat** di [wichat-system-design.md](wichat-system-design.md).

Permissions dibagi menjadi dua tier:

- **Core (Active / Immediate):** Digunakan langsung pada fitur inti WiChat saat ini & fase awal (Auth/Group, Channels, Chat, Moderation, LiveKit).
- **Next Feature (Future-Ready):** Disiapkan untuk fase lanjutan sesuai arsitektur (`ms-audit`, `ms-plugin`, Threads/Forum, Polls, Custom Expressions di MinIO).

---

## 1. General Group Permissions

| Permission Key        | Label (WiChat)          | Deskripsi                                                                                                                                                                        | Tier         | Catatan Implementasi                                                 |
| :-------------------- | :---------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :----------- | :------------------------------------------------------------------- |
| `view_channels`       | **View Channels**       | Allows members to view channels by default (excluding private channels).                                                                                                         | Core         | Default ON di `@everyone`.                                           |
| `manage_channels`     | **Manage Channels**     | Allows members to create, edit, or delete channels and categories.                                                                                                               | Core         | Mencakup pengaturan private channel.                                 |
| `manage_roles`        | **Manage Roles**        | Allows members to create new roles and edit or delete roles lower than their highest role. Also allows members to change permissions of individual channels they have access to. | Core         | Menghormati hierarki posisi role.                                    |
| `manage_group`        | **Manage Group**        | Allows members to change this group's name, icon, banner, and general group settings.                                                                                            | Core         | Adaptasi dari _Manage Server_ (tanpa switch region/AutoMod Discord). |
| `create_expressions`  | **Create Expressions**  | Allows members to add custom emoji, stickers, and sounds in this group.                                                                                                          | Next Feature | Aset disimpan di bucket publik MinIO.                                |
| `manage_expressions`  | **Manage Expressions**  | Allows members to edit or remove custom emoji, stickers, and sounds in this group.                                                                                               | Next Feature | Aset disimpan di bucket publik MinIO.                                |
| `view_audit_log`      | **View Audit Log**      | Allows members to view a record of who made which changes in this group.                                                                                                         | Next Feature | Disuplai oleh layanan `ms-audit` (Kafka → Postgres).                 |
| `view_group_insights` | **View Group Insights** | Allows members to view Group Insights, showing data on member growth, engagement, and channel activity.                                                                          | Next Feature | Disuplai oleh analytics consumer dari Kafka.                         |
| `manage_webhooks`     | **Manage Webhooks**     | Allows members to create, edit, or delete webhooks that can post messages from external tools into channels.                                                                     | Next Feature | Ditunda di v1 (§3 Non-Goals), masuk saat integrasi webhook dibuka.   |

---

## 2. Membership Permissions

| Permission Key     | Label (WiChat)                        | Deskripsi                                                                                                                                            | Tier         | Catatan Implementasi                                         |
| :----------------- | :------------------------------------ | :--------------------------------------------------------------------------------------------------------------------------------------------------- | :----------- | :----------------------------------------------------------- |
| `create_invite`    | **Create Invite**                     | Allows members to invite new people to this group.                                                                                                   | Core         | Mengontrol pembuatan/penggunaan invite link & dialog invite. |
| `change_nickname`  | **Change Nickname**                   | Allows members to change their own nickname, a custom name for just this group.                                                                      | Next Feature | Saat ini memakai `display_name` global akun.                 |
| `manage_nicknames` | **Manage Nicknames**                  | Allows members to change the nicknames of other members.                                                                                             | Next Feature | Berpasangan dengan `change_nickname`.                        |
| `kick_members`     | **Kick, Approve, and Reject Members** | Kick removes other members from this group (they can rejoin with a new invite). If approval is enabled, allows approving or rejecting join requests. | Core         | Sudah menjadi bagian baseline moderation (§6.8).             |
| `ban_members`      | **Ban Members**                       | Allows members to permanently ban and delete the message history of other members from this group.                                                   | Core         | Sudah menjadi bagian baseline moderation (§6.8).             |
| `timeout_members`  | **Timeout Members**                   | Timed-out members cannot send messages in chat, reply within threads, react to messages, or speak in voice channels.                                 | Next Feature | Moderasi sementara (mute sementara di seluruh group).        |

---

## 3. Text & Forum Channel Permissions

| Permission Key             | Label (WiChat)                                  | Deskripsi                                                                                                                     | Tier         | Catatan Implementasi                                  |
| :------------------------- | :---------------------------------------------- | :---------------------------------------------------------------------------------------------------------------------------- | :----------- | :---------------------------------------------------- |
| `send_messages`            | **Send Messages and Create Posts**              | Allows members to send messages in text channels and create posts in forum channels.                                          | Core         | Default ON di `@everyone`.                            |
| `send_messages_in_threads` | **Send Messages in Threads and Posts**          | Allows members to send messages in threads and in posts on forum channels.                                                    | Next Feature | Dibutuhkan saat fitur Threads & Forum channels aktif. |
| `create_public_threads`    | **Create Public Threads**                       | Allows members to create threads that everyone in a channel can view.                                                         | Next Feature | Fitur thread `ms-chat`.                               |
| `create_private_threads`   | **Create Private Threads**                      | Allows members to create invite-only threads.                                                                                 | Next Feature | Fitur thread `ms-chat`.                               |
| `embed_links`              | **Embed Links**                                 | Allows links that members share to show rich embedded previews in text channels.                                              | Next Feature | Untuk link preview card.                              |
| `attach_files`             | **Attach Files**                                | Allows members to upload files or media in text channels.                                                                     | Core         | Upload langsung ke MinIO via presigned URL.           |
| `add_reactions`            | **Add Reactions**                               | Allows members to add new emoji reactions to a message. If disabled, members can still toggle existing reactions.             | Core         | Sudah ada pada UI reaksi pesan `ms-chat`.             |
| `mention_everyone`         | **Mention `@everyone`, `@here`, and All Roles** | Allows members to use `@everyone` or `@here`, and `@mention` all roles even if a role's mentionable toggle is disabled.       | Core         | Terhubung dengan `ms-notification`.                   |
| `manage_messages`          | **Manage Messages**                             | Allows members to delete or remove embeds from messages sent by other members.                                                | Core         | Baseline moderation (§6.8).                           |
| `pin_messages`             | **Pin Messages**                                | Allows members to pin or unpin any message in a channel.                                                                      | Core         | Tab/fitur Pins di `ms-chat`.                          |
| `bypass_slowmode`          | **Bypass Slowmode**                             | Allows members to send messages without being affected by channel slowmode.                                                   | Next Feature | Moderasi & anti-spam channel.                         |
| `manage_threads`           | **Manage Threads and Posts**                    | Allows members to rename, delete, close, and turn on slowmode for threads and posts, as well as view private threads.         | Next Feature | Moderasi thread & forum.                              |
| `read_message_history`     | **Read Message History**                        | Allows members to read previous messages sent in channels. If disabled, members only see messages sent while they are online. | Core         | Mengontrol akses infinite scroll history `ms-chat`.   |
| `send_voice_messages`      | **Send Voice Messages**                         | Allows members to record and send voice notes in text channels.                                                               | Next Feature | Media audio via MinIO.                                |
| `create_polls`             | **Create Polls**                                | Allows members to create polls in channels.                                                                                   | Next Feature | Bisa dibangun sebagai core message type atau plugin.  |

---

## 4. Voice & Video Channel Permissions (LiveKit)

| Permission Key       | Label (WiChat)           | Deskripsi                                                                                               | Tier         | Catatan Implementasi                                  |
| :------------------- | :----------------------- | :------------------------------------------------------------------------------------------------------ | :----------- | :---------------------------------------------------- |
| `connect`            | **Connect**              | Allows members to join voice and video channels and hear others.                                        | Core         | Mengontrol penerbitan token room LiveKit (`ms-core`). |
| `speak`              | **Speak**                | Allows members to talk in voice channels. If disabled, members join muted until unmuted by a moderator. | Core         | LiveKit audio publish permission.                     |
| `video`              | **Video & Screen Share** | Allows members to share their camera video or screen share in voice channels.                           | Core         | LiveKit video/screen track publish permission.        |
| `use_voice_activity` | **Use Voice Activity**   | Allows members to speak automatically without Push-to-Talk. If disabled, Push-to-Talk is required.      | Next Feature | Pengaturan client audio input.                        |
| `use_soundboard`     | **Use Soundboard**       | Allows members to play sounds from the group soundboard in voice channels.                              | Next Feature | Aset soundboard di MinIO.                             |
| `priority_speaker`   | **Priority Speaker**     | Allows members to be heard more clearly by lowering others' volume when speaking.                       | Next Feature | Ducking audio client-side / LiveKit.                  |
| `mute_members`       | **Mute Members**         | Allows members to server-mute other members in voice channels.                                          | Core         | Baseline moderation (§6.8) via `ms-core` → LiveKit.   |
| `deafen_members`     | **Deafen Members**       | Allows members to server-deafen other members in voice channels.                                        | Core         | Baseline moderation (§6.8) via `ms-core` → LiveKit.   |
| `move_members`       | **Move Members**         | Allows members to disconnect or move other members between voice channels they have access to.          | Next Feature | Kontrol perpindahan room LiveKit.                     |

---

## 5. Plugins & Events Permissions

> **Catatan Arsitektur WiChat:** WiChat menggunakan arsitektur plugin proses terpisah (`ms-plugin` via gRPC). Selain permission manajemen plugin inti di bawah, setiap plugin yang di-install dapat mendaftarkan permission spesifiknya sendiri ke tabel `permissions` (contoh: `taskboard.create_card`, `whiteboard.edit`).

| Permission Key   | Label (WiChat)             | Deskripsi                                                                                                              | Tier          | Catatan Implementasi                                                        |
| :--------------- | :------------------------- | :--------------------------------------------------------------------------------------------------------------------- | :------------ | :-------------------------------------------------------------------------- |
| `manage_plugins` | **Manage Plugins**         | Allows members to install, configure, enable, or remove plugins and approve plugin permission manifests in this group. | Core (Plugin) | Pengganti konsep _Apps Permissions_ Discord untuk arsitektur plugin WiChat. |
| `use_plugins`    | **Use Plugins & Commands** | Allows members to interact with enabled group plugins and invoke plugin commands.                                      | Core (Plugin) | Kontrol akses penggunaan plugin di dalam channel.                           |
| `create_events`  | **Create Events**          | Allows members to create scheduled group events.                                                                       | Next Feature  | Bisa diimplementasikan via plugin Calendar/Events.                          |
| `manage_events`  | **Manage Events**          | Allows members to edit and cancel scheduled group events.                                                              | Next Feature  | Moderasi event terjadwal.                                                   |

---

## 6. Advanced Permissions

| Permission Key  | Label (WiChat)    | Deskripsi                                                                                                                                             | Tier | Catatan Implementasi                                                                     |
| :-------------- | :---------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------- | :--- | :--------------------------------------------------------------------------------------- |
| `administrator` | **Administrator** | Members with this permission have every permission and bypass all channel-specific restrictions (including private channels). **Grant with caution.** | Core | Untuk role admin/co-owner tanpa memindahkan kepemilikan grup (`Owner` tetap badge-only). |

---

## Ringkasan: Permissions Discord yang Tidak Dipakai di WiChat

| Permission Discord                         | Alasan Dihilangkan                                                                                                                           |
| :----------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------- |
| **Use External Emoji / Stickers / Sounds** | Bergantung pada konsep langganan SaaS/Nitro lintas server; WiChat adalah platform _self-hosted single-tenant_.                               |
| **Use Activities / Use External Apps**     | Digantikan oleh model plugin terkurasi WiChat (`manage_plugins`, `use_plugins`, dan plugin-registered permissions).                          |
| **Send Text-to-Speech Messages (`/tts`)**  | Tidak relevan untuk kebutuhan komunikasi tim inti WiChat.                                                                                    |
| **Request to Speak (Stage Channel)**       | Fitur Stage Channel tidak masuk dalam cakupan tipe channel WiChat (`broadcast`, `text_only`, `voice_only`, `text_and_voice`, `forum`, `dm`). |
| **Set Voice Channel Status**               | Detail kosmetik minor yang tidak diperlukan di arsitektur voice WiChat saat ini.                                                             |
