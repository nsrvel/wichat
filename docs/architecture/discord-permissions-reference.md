# Discord Role Permissions Reference (Full Extraction)

Dokumen ini berisi daftar lengkap perizinan role (permissions) yang ada di Discord beserta kategori, nama permission, dan deskripsinya sesuai tangkapan layar referensi.

---

## 1. General Server Permissions

| Permission               | Deskripsi                                                                                                                                                                                                  |
| :----------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **View Channels**        | Allows members to view channels by default (excluding private channels).                                                                                                                                   |
| **Manage Channels**      | Allows members to create, edit, or delete channels.                                                                                                                                                        |
| **Manage Roles**         | Allows members to create new roles and edit or delete roles lower than their highest role. Also allows members to change permissions of individual channels that they have access to.                      |
| **Create Expressions**   | Allows members to add custom emoji, stickers, and sounds in this server.                                                                                                                                   |
| **Manage Expressions**   | Allows members to edit or remove custom emoji, stickers, and sounds in this server.                                                                                                                        |
| **View Audit Log**       | Allows members to view a record of who made which changes in this server.                                                                                                                                  |
| **View Server Insights** | Allows members to view Server Insights, which shows data on community growth, engagement, and more. This will allow them to see certain data about channel activity, even for channels they cannot access. |
| **Manage Webhooks**      | Allows members to create, edit, or delete webhooks, which can post messages from other apps or sites into this server.                                                                                     |
| **Manage Server**        | Allow members to change this server's name, switch regions, view all invites, add apps to this server and create and update AutoMod rules.                                                                 |

---

## 2. Membership Permissions

| Permission                            | Deskripsi                                                                                                                                                                                                                                            |
| :------------------------------------ | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Create Invite**                     | Allows members to invite new people to this server.                                                                                                                                                                                                  |
| **Change Nickname**                   | Allows members to change their own nickname, a custom name for just this server.                                                                                                                                                                     |
| **Manage Nicknames**                  | Allows members to change the nicknames of other members.                                                                                                                                                                                             |
| **Kick, Approve, and Reject Members** | Kick will remove other members from this server. Kicked members will be able to rejoin if they have another invite. If the server enables Member Requirements, this permission enables the ability to approve or reject members who request to join. |
| **Ban Members**                       | Allows members to permanently ban and delete the message history of other members from this server.                                                                                                                                                  |
| **Timeout Members**                   | When you put a user in timeout they will not be able to send messages in chat, reply within threads, react to messages, or speak in voice or Stage channels.                                                                                         |

---

## 3. Text Channel Permissions

| Permission                                  | Deskripsi                                                                                                                                                                                                                 |
| :------------------------------------------ | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Send Messages and Create Posts**          | Allow members to send messages in text channels and create posts in forum channels.                                                                                                                                       |
| **Send Messages in Threads and Posts**      | Allow members to send messages in threads and in posts on forum channels.                                                                                                                                                 |
| **Create Public Threads**                   | Allow members to create threads that everyone in a channel can view.                                                                                                                                                      |
| **Create Private Threads**                  | Allow members to create invite-only threads.                                                                                                                                                                              |
| **Embed Links**                             | Allows links that members share to show embedded content in text channels.                                                                                                                                                |
| **Attach Files**                            | Allows members to upload files or media in text channels.                                                                                                                                                                 |
| **Add Reactions**                           | Allows members to add new emoji reactions to a message. If this permission is disabled, members can still react using any existing reactions on a message.                                                                |
| **Use External Emoji**                      | Allows members to use emoji from other servers, if they're a Discord Nitro member.                                                                                                                                        |
| **Use External Stickers**                   | Allows members to use stickers from other servers, if they're a Discord Nitro member.                                                                                                                                     |
| **Mention @everyone, @here, and All Roles** | Allows members to use @everyone (everyone in the server) or @here (only online members in that channel). They can also @mention all roles, even if the role's "Allow anyone to mention this role" permission is disabled. |
| **Manage Messages**                         | Allows members to delete or remove embeds from messages by other members.                                                                                                                                                 |
| **Pin Messages**                            | Allows members to pin or unpin any message.                                                                                                                                                                               |
| **Bypass Slowmode**                         | Allows members to send messages without being affected by slowmode.                                                                                                                                                       |
| **Manage Threads and Posts**                | Allows members to rename, delete, close, and turn on slow mode for threads and posts. They can also view private threads.                                                                                                 |
| **Read Message History**                    | Allows members to read previous messages sent in channels. If this permission is disabled, members only see messages sent when they are online. This does not fully apply to threads and forum posts.                     |
| **Send Text-to-Speech Messages**            | Allows members to send text-to-speech messages by starting a message with `/tts`. These messages can be heard by anyone focused on the channel.                                                                           |
| **Send Voice Messages**                     | Allows members to send voice messages.                                                                                                                                                                                    |
| **Create Polls**                            | Allows members to create polls.                                                                                                                                                                                           |

---

## 4. Voice Channel Permissions

| Permission                   | Deskripsi                                                                                                                                                                                                                         |
| :--------------------------- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Connect**                  | Allows members to join voice channels and hear others.                                                                                                                                                                            |
| **Speak**                    | Allows members to talk in voice channels. If this permission is disabled, members are default muted until somebody with the "Mute Members" permission un-mutes them.                                                              |
| **Video**                    | Allows members to share their video, screen share, or stream a game in this server.                                                                                                                                               |
| **Use Soundboard**           | Allows members to send sounds from server soundboard.                                                                                                                                                                             |
| **Use External Sounds**      | Allows members to use sounds from other servers, if they're a Discord Nitro member.                                                                                                                                               |
| **Use Voice Activity**       | Allows members to speak in voice channels by simply talking. If this permission is disabled, members are required to use Push-to-talk. Good for controlling background noise or noisy members.                                    |
| **Priority Speaker**         | Allows members to be more easily heard in voice channels. When activated, the volume of others without this permission will be automatically lowered. Priority Speaker is activated by using the Push to Talk (Priority) keybind. |
| **Mute Members**             | Allows members to mute other members in voice channels for everyone.                                                                                                                                                              |
| **Deafen Members**           | Allows members to deafen other members in voice channels, which means they won't be able to speak or hear others.                                                                                                                 |
| **Move Members**             | Allows members to disconnect or move other members between voice channels that the member with this permission has access to.                                                                                                     |
| **Set Voice Channel Status** | Allows members to create and edit voice channel status.                                                                                                                                                                           |

---

## 5. Apps Permissions

| Permission                   | Deskripsi                                                                                                           |
| :--------------------------- | :------------------------------------------------------------------------------------------------------------------ |
| **Use Application Commands** | Allows members to use commands from applications, including slash commands and context menu commands.               |
| **Use Activities**           | Allows members to use Activities.                                                                                   |
| **Use External Apps**        | Allows apps that members have added to their account to post messages. When disabled, the messages will be private. |

---

## 6. Stage Channel Permissions

| Permission           | Deskripsi                                                                                          |
| :------------------- | :------------------------------------------------------------------------------------------------- |
| **Request to Speak** | Allow requests to speak in Stage channels. Stage moderators manually approve or deny each request. |

---

## 7. Events Permissions

| Permission        | Deskripsi                                 |
| :---------------- | :---------------------------------------- |
| **Create Events** | Allows members to create events.          |
| **Manage Events** | Allows members to edit and cancel events. |

---

## 8. Advanced Permissions

| Permission        | Deskripsi                                                                                                                                                                                                                                         |
| :---------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Administrator** | Members with this permission will have every permission and will also bypass all channel specific permissions or restrictions (for example, these members would get access to all private channels). **This is a dangerous permission to grant.** |
