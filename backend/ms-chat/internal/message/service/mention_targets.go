package service

import (
	"context"
	"strings"

	"github.com/wichat/wichat/backend/ms-chat/pkg/mention"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func roleMentionAllowed(role *userv1.GroupRole, callerPerms []string) bool {
	if role == nil {
		return false
	}
	if role.GetIsSystem() && strings.EqualFold(strings.TrimSpace(role.GetSlug()), "everyone") {
		return hasAnyPerm(callerPerms, "administrator", "mention_everyone")
	}
	return role.GetMentionable() || hasAnyPerm(callerPerms, "administrator", "mention_everyone")
}

func (s service) resolveMentionPublicIDs(
	ctx context.Context,
	accessToken, groupPublicID, callerPublicID string,
	callerPerms []string,
	body string,
	explicit []string,
) ([]string, error) {
	userIDs := mention.StripCallerPublicID(mention.MergePublicUserIDs(explicit, body), callerPublicID)
	roleIDs := mention.PublicRoleIDsFromBody(body)
	if len(roleIDs) == 0 || s.groupRoles == nil {
		return userIDs, nil
	}

	listResp, err := s.groupRoles.ListGroupRoles(ctx, &userv1.ListGroupRolesRequest{
		AccessToken: accessToken,
		GroupId:     groupPublicID,
	})
	if err != nil {
		return nil, err
	}

	roleByID := make(map[string]*userv1.GroupRole, len(listResp.GetRoles()))
	for _, r := range listResp.GetRoles() {
		if r == nil {
			continue
		}
		id := strings.TrimSpace(r.GetId())
		if id != "" {
			roleByID[id] = r
		}
	}

	fromRoles := make([]string, 0)
	for _, rolePublicID := range roleIDs {
		rolePublicID = strings.TrimSpace(rolePublicID)
		role, ok := roleByID[rolePublicID]
		if !ok {
			return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "mention_role"})
		}
		if !roleMentionAllowed(role, callerPerms) {
			return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
		}

		membersResp, err := s.groupRoles.ListGroupRoleMembers(ctx, &userv1.ListGroupRoleMembersRequest{
			AccessToken: accessToken,
			GroupId:     groupPublicID,
			RoleId:      rolePublicID,
		})
		if err != nil {
			return nil, err
		}
		for _, m := range membersResp.GetMembers() {
			if m == nil {
				continue
			}
			uid := strings.TrimSpace(m.GetUserId())
			if uid != "" && uid != callerPublicID {
				fromRoles = append(fromRoles, uid)
			}
		}
	}

	return mention.StripCallerPublicID(mention.UnionPublicUserIDs(userIDs, fromRoles), callerPublicID), nil
}
