import {
	CategoryPermissionsListSchema,
	CreateChannelCategoryRequestSchema,
	CreateChannelCategoryResponseSchema,
	ReorderChannelCategoriesRequestSchema,
	UpdateChannelCategoryRequestSchema,
	UpdateChannelCategoryResponseSchema,
	UpsertCategoryPermissionOverrideRequestSchema,
	UpsertCategoryPermissionOverrideResponseSchema,
	type CategoryPermissionTargetType,
	type CategoryPermissionsList,
	type CreateChannelCategoryRequest,
	type UpdateChannelCategoryRequest,
	type UpsertCategoryPermissionOverrideRequest,
} from "@wichat/contracts/chat/channel-category";
import { OkSchema } from "@wichat/contracts/common";
import { api } from "@/integrations/api-client";

export async function createGroupCategory(groupId: string, payload: CreateChannelCategoryRequest) {
	const body = CreateChannelCategoryRequestSchema.parse(payload);
	const data = await api.post(
		`/groups/${encodeURIComponent(groupId)}/categories`,
		CreateChannelCategoryResponseSchema,
		body,
	);
	return data.category;
}

export async function updateGroupCategory(
	groupId: string,
	categoryId: string,
	payload: UpdateChannelCategoryRequest,
) {
	const body = UpdateChannelCategoryRequestSchema.parse(payload);
	const data = await api.patch(
		`/groups/${encodeURIComponent(groupId)}/categories/${encodeURIComponent(categoryId)}`,
		UpdateChannelCategoryResponseSchema,
		body,
	);
	return data.category;
}

export async function deleteGroupCategory(groupId: string, categoryId: string) {
	return api.delete(
		`/groups/${encodeURIComponent(groupId)}/categories/${encodeURIComponent(categoryId)}`,
		OkSchema,
	);
}

export async function listCategoryPermissions(
	groupId: string,
	categoryId: string,
): Promise<CategoryPermissionsList> {
	return api.get(
		`/groups/${encodeURIComponent(groupId)}/categories/${encodeURIComponent(categoryId)}/permissions`,
		CategoryPermissionsListSchema,
	);
}

export async function upsertCategoryPermission(
	groupId: string,
	categoryId: string,
	payload: UpsertCategoryPermissionOverrideRequest,
) {
	const body = UpsertCategoryPermissionOverrideRequestSchema.parse(payload);
	const data = await api.put(
		`/groups/${encodeURIComponent(groupId)}/categories/${encodeURIComponent(categoryId)}/permissions`,
		UpsertCategoryPermissionOverrideResponseSchema,
		body,
	);
	return data.override;
}

export async function deleteCategoryPermission(
	groupId: string,
	categoryId: string,
	targetType: CategoryPermissionTargetType,
	targetId: string,
) {
	return api.delete(
		`/groups/${encodeURIComponent(groupId)}/categories/${encodeURIComponent(categoryId)}/permissions/${encodeURIComponent(targetType)}/${encodeURIComponent(targetId)}`,
		OkSchema,
	);
}

export async function reorderGroupCategories(groupId: string, categoryIds: string[]) {
	const body = ReorderChannelCategoriesRequestSchema.parse({ category_ids: categoryIds });
	return api.put(`/groups/${encodeURIComponent(groupId)}/categories/order`, OkSchema, body);
}
