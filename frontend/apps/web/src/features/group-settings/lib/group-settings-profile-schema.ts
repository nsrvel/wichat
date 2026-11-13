import { groupNameSchema } from "@wichat/contracts";
import { z } from "zod";

export const groupSettingsProfileSchema = z.object({
	name: groupNameSchema,
	description: z.string().trim().max(2000),
	is_private: z.boolean(),
});

export type GroupSettingsProfileFormValues = z.infer<typeof groupSettingsProfileSchema>;
