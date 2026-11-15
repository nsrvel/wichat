import {
	AddGroupRoleMembersRequestSchema,
	CreateGroupRoleRequestSchema,
	GroupRoleMembersResponseSchema,
	GroupRoleResponseSchema,
	GroupRolesListResponseSchema,
	UpdateGroupRoleRequestSchema,
	type AddGroupRoleMembersRequest,
	type CreateGroupRoleRequest,
	type UpdateGroupRoleRequest,
} from "@wichat/contracts/group/role";
import { OkSchema } from "@wichat/contracts/common";
import {
	UpdateGroupHideRolesRequestSchema,
	UpdateGroupHideRolesResponseSchema,
} from "@wichat/contracts/group";
import { api } from "@/integrations/api-client";

function rolesBase(groupId: string) {
	return `/groups/${encodeURIComponent(groupId)}/roles`;
}

export async function listGroupRoles(groupId: string) {
	return api.get(rolesBase(groupId), GroupRolesListResponseSchema);
}

export async function createGroupRole(groupId: string, payload: CreateGroupRoleRequest = { name: "new role" }) {
	const body = CreateGroupRoleRequestSchema.parse(payload);
	const res = await api.post(rolesBase(groupId), GroupRoleResponseSchema, body);
	return res.role;
}

export async function updateGroupRole(groupId: string, roleId: string, payload: UpdateGroupRoleRequest) {
	const body = UpdateGroupRoleRequestSchema.parse(payload);
	const res = await api.patch(`${rolesBase(groupId)}/${encodeURIComponent(roleId)}`, GroupRoleResponseSchema, body);
	return res.role;
}

export async function duplicateGroupRole(groupId: string, roleId: string) {
	const res = await api.post(
		`${rolesBase(groupId)}/${encodeURIComponent(roleId)}/duplicate`,
		GroupRoleResponseSchema,
		{},
	);
	return res.role;
}

export async function deleteGroupRole(groupId: string, roleId: string) {
	return api.delete(`${rolesBase(groupId)}/${encodeURIComponent(roleId)}`, OkSchema);
}

export async function listGroupRoleMembers(groupId: string, roleId: string) {
	const res = await api.get(
		`${rolesBase(groupId)}/${encodeURIComponent(roleId)}/members`,
		GroupRoleMembersResponseSchema,
	);
	return res.members;
}

export async function addGroupRoleMembers(
	groupId: string,
	roleId: string,
	payload: AddGroupRoleMembersRequest,
) {
	const body = AddGroupRoleMembersRequestSchema.parse(payload);
	const res = await api.post(
		`${rolesBase(groupId)}/${encodeURIComponent(roleId)}/members`,
		GroupRoleResponseSchema,
		body,
	);
	return res.role;
}

export async function removeGroupRoleMember(groupId: string, roleId: string, userId: string) {
	const res = await api.delete(
		`${rolesBase(groupId)}/${encodeURIComponent(roleId)}/members/${encodeURIComponent(userId)}`,
		GroupRoleResponseSchema,
	);
	return res.role;
}

export async function updateGroupHideRoles(groupId: string, hideRoles: boolean) {
	const body = UpdateGroupHideRolesRequestSchema.parse({ hide_roles: hideRoles });
	const res = await api.patch(`${rolesBase(groupId)}/settings`, UpdateGroupHideRolesResponseSchema, body);
	return res.hide_roles;
}
