import { z } from "zod";

export type UserRecipientChip = {
	kind: "user";
	id: string;
	label: string;
	email: string;
};

export type EmailRecipientChip = {
	kind: "email";
	email: string;
	label: string;
};

export type RecipientChip = UserRecipientChip | EmailRecipientChip;

const emailSchema = z.email();

export function parseEmailChip(raw: string): EmailRecipientChip | null {
	const trimmed = raw.trim();
	if (!trimmed) {
		return null;
	}
	if (!emailSchema.safeParse(trimmed).success) {
		return null;
	}
	return { kind: "email", email: trimmed.toLowerCase(), label: trimmed };
}
