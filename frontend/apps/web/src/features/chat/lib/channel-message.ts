export type ChannelMessageReaction = {
	emoji: string;
	count?: number;
	active?: boolean;
};

export type ChannelMessage = {
	id: string;
	authorId: string;
	authorDisplayName: string;
	authorAvatarUrl: string;
	/** ISO-8601 from API later; dev fixtures may use display strings via `timeLabel`. */
	createdAt: string;
	bodyPlain: string;
	mine?: boolean;
	reactions?: ChannelMessageReaction[];
	/** Relative or formatted time for display when API formatting is not wired. */
	timeLabel?: string;
};
