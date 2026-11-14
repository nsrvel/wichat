export type ExpressionTab = "all" | "emoji" | "gifs" | "stickers";

export type ExpressionKind = "emoji" | "gif" | "sticker";

export type MediaExpressionKind = "gifs" | "stickers";

export type EmojiCategoryId =
	| "smileys"
	| "people"
	| "nature"
	| "food"
	| "travel"
	| "activities"
	| "objects"
	| "symbols"
	| "flags";

export type EmojiEntry = {
	hexcode: string;
	unicode: string;
	label: string;
	category: EmojiCategoryId;
	tags: string[];
	skins?: Array<{
		hexcode: string;
		unicode: string;
		label: string;
		tone: number | number[];
	}>;
};

export type MediaExpressionEntry = {
	id: string;
	title: string;
	url: string;
	previewUrl: string;
	width: number;
	height: number;
	provider: "klipy" | "giphy";
};

export type GifEntry = MediaExpressionEntry;

export type StickerEntry = MediaExpressionEntry;

export type MediaExpressionCategory = {
	id: string;
	label: string;
	query: string;
	previewUrl: string;
};

export type GifCategory = MediaExpressionCategory;

export type StickerCategory = MediaExpressionCategory;

export type ExpressionSelection =
	| {
			kind: "emoji";
			emoji: string;
			name: string;
			shortcode?: string;
	  }
	| {
			kind: "gif";
			id: string;
			title: string;
			url: string;
			previewUrl: string;
			width: number;
			height: number;
			provider: "klipy" | "giphy";
	  }
	| {
			kind: "sticker";
			id: string;
			title: string;
			url: string;
			previewUrl: string;
			width: number;
			height: number;
			provider: "klipy" | "giphy";
	  };
