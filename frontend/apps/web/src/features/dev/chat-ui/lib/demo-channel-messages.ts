export const demoAvatar = {
	self: "/avatars/demo/marvell.png",
	yadi: "/avatars/demo/koh-yadi.jpg",
	yanto: "/avatars/demo/koh-yanto.jpg",
} as const;

export const demoName = {
	self: "ⱮąɾѵҽӀӀ",
	yadi: "Koh Yadi",
	yanto: "Koh Yanto",
} as const;

export const demoMentionHandles = [demoName.yadi, demoName.yanto, demoName.self] as const;
