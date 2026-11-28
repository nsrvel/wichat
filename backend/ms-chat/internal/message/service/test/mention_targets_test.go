package service_test

import (
	"context"
	"testing"

	mentionmock "github.com/wichat/wichat/backend/ms-chat/internal/mention/repository/mock"
	"github.com/wichat/wichat/backend/ms-chat/internal/message/repository/mock"
	"github.com/wichat/wichat/backend/ms-chat/internal/message/service"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	outboxmock "github.com/wichat/wichat/backend/ms-chat/internal/outbox/repository/mock"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	grouprole "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/group-role"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

type fakeGroupRoles struct {
	grouprole.Client
	roles   []*userv1.GroupRole
	members map[string][]*userv1.GroupMemberEntry
}

func (f fakeGroupRoles) ListGroupRoles(context.Context, *userv1.ListGroupRolesRequest) (*userv1.ListGroupRolesResponse, error) {
	return &userv1.ListGroupRolesResponse{Roles: f.roles}, nil
}

func (f fakeGroupRoles) ListGroupRoleMembers(_ context.Context, req *userv1.ListGroupRoleMembersRequest) (*userv1.ListGroupRoleMembersResponse, error) {
	return &userv1.ListGroupRoleMembersResponse{Members: f.members[req.GetRoleId()]}, nil
}

func TestCreateChannelMessageResolvesRoleMentionsAndChecksPermissions(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	roleMentionableID := "0123456789abcdef01234567"
	roleEveryoneID := "0123456789abcdef01234568"
	roleHiddenID := "0123456789abcdef01234569"

	rolesClient := fakeGroupRoles{
		roles: []*userv1.GroupRole{
			{Id: roleMentionableID, Slug: "mods", Name: "Mods", Mentionable: true},
			{Id: roleEveryoneID, Slug: "everyone", Name: "everyone", IsSystem: true},
			{Id: roleHiddenID, Slug: "staff", Name: "Staff", Mentionable: false},
		},
		members: map[string][]*userv1.GroupMemberEntry{
			roleMentionableID: {
				{UserId: "u1"},
				{UserId: "u2"},
				{UserId: "u3"},
			},
			roleEveryoneID: {
				{UserId: "u1"},
				{UserId: "u2"},
			},
		},
	}

	st := mock.NewState()
	st.Authors["u1"] = "Alice"
	mState := mentionmock.NewState()
	mRepo := mentionmock.NewPgRepositoryMock(mState)

	svcMember := service.NewService(
		mock.NewPgRepositoryMock(st),
		mRepo,
		nil,
		nil,
		fakeChannel{perms: []string{"view_channels", "send_messages"}},
		rolesClient,
		outboxmock.NewPgRepositoryMock(outboxmock.NewState()),
		nil,
	)

	// 1. Mentionable role expands to u2, u3 and excludes author u1.
	msg, err := svcMember.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		GroupPublicID:   "g1",
		ChannelPublicID: "ch1",
		AccessToken:     testAccessToken("u1"),
		Body:            "hi [@Mods](wichat-mention-role:" + roleMentionableID + ")",
	})
	if err != nil {
		t.Fatalf("unexpected err for mentionable role: %v", err)
	}
	if len(msg.MentionedUserIDs) != 2 || msg.MentionedUserIDs[0] != "u2" || msg.MentionedUserIDs[1] != "u3" {
		t.Fatalf("expected [u2 u3], got %v", msg.MentionedUserIDs)
	}

	// 2. Regular member cannot mention @everyone or non-mentionable role.
	for _, rid := range []string{roleEveryoneID, roleHiddenID} {
		_, err = svcMember.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
			GroupPublicID:   "g1",
			ChannelPublicID: "ch1",
			AccessToken:     testAccessToken("u1"),
			Body:            "ping [@role](wichat-mention-role:" + rid + ")",
		})
		if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindPermissionDenied {
			t.Fatalf("expected permission denied for role %s, got %v", rid, err)
		}
	}

	// 3. Caller with mention_everyone permission can mention @everyone.
	svcMod := service.NewService(
		mock.NewPgRepositoryMock(st),
		mRepo,
		nil,
		nil,
		fakeChannel{perms: []string{"view_channels", "send_messages", "mention_everyone"}},
		rolesClient,
		outboxmock.NewPgRepositoryMock(outboxmock.NewState()),
		nil,
	)
	msgEveryone, err := svcMod.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		GroupPublicID:   "g1",
		ChannelPublicID: "ch1",
		AccessToken:     testAccessToken("u1"),
		Body:            "hi [@everyone](wichat-mention-role:" + roleEveryoneID + ")",
	})
	if err != nil {
		t.Fatalf("unexpected err for @everyone with perm: %v", err)
	}
	if len(msgEveryone.MentionedUserIDs) != 1 || msgEveryone.MentionedUserIDs[0] != "u2" {
		t.Fatalf("expected [u2], got %v", msgEveryone.MentionedUserIDs)
	}
}
