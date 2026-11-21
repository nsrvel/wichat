package authHttp_test

import (
	"testing"

	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func TestCategoryPermissionOverrideFromPB(t *testing.T) {
	nilBody := authHttp.CategoryPermissionOverrideFromPB(nil)
	if len(nilBody.AllowPermissions) != 0 {
		t.Fatal("expected empty allow slice")
	}
	body := authHttp.CategoryPermissionOverrideFromPB(&userv1.ChannelCategoryPermissionOverride{
		TargetType:       "role",
		TargetId:         "r1",
		AllowPermissions: []string{"read"},
		DenyPermissions:  []string{"write"},
	})
	if body.TargetID != "r1" || len(body.AllowPermissions) != 1 {
		t.Fatalf("%+v", body)
	}
}

func TestGroupAndChannelMappers(t *testing.T) {
	if authHttp.GroupFromPB(nil).ID != "" {
		t.Fatal("nil group")
	}
	g := authHttp.GroupFromPB(&userv1.GroupSummary{Id: "g1", Name: "WiChat", EffectivePermissions: nil})
	if g.ID != "g1" || len(g.EffectivePermissions) != 0 {
		t.Fatalf("%+v", g)
	}

	ch := authHttp.ChannelFromPB(&userv1.ChannelSummary{Id: "c1", Name: "general", Type: "text"})
	if ch.ID != "c1" {
		t.Fatalf("%+v", ch)
	}
	if authHttp.ChannelFromPB(nil).ID != "" {
		t.Fatal("nil channel")
	}

	cat := authHttp.ChannelCategoryFromPB(&userv1.ChannelCategory{Id: "cat-1", Name: "Text"})
	if cat.ID != "cat-1" {
		t.Fatal(cat)
	}
	cats2 := authHttp.ChannelCategoriesFromPB([]*userv1.ChannelCategory{{Id: "cat-1", Name: "Text"}})
	if len(cats2) != 1 {
		t.Fatal(cats2)
	}

	chans := authHttp.ChannelsFromPB([]*userv1.ChannelSummary{{Id: "c1", Name: "general"}})
	if len(chans) != 1 {
		t.Fatal(chans)
	}

	groups := authHttp.GroupsFromPB([]*userv1.GroupSummary{{Id: "g1", Name: "G"}})
	if len(groups) != 1 {
		t.Fatal(groups)
	}
}

func TestGroupMembersAndSearchMappers(t *testing.T) {
	members := authHttp.GroupMembersFromPB([]*userv1.GroupMemberEntry{
		{UserId: "u1", DisplayName: "Ada", Roles: []*userv1.GroupMemberRoleSummary{{Id: "r1", Name: "Admin"}}},
		nil,
	})
	if len(members) != 2 || members[0].UserID != "u1" {
		t.Fatalf("%+v", members)
	}

	search := authHttp.UserSearchFromPB([]*userv1.UserSearchHit{{UserId: "u1", Username: "ada"}})
	if len(search) != 1 || search[0].UserID != "u1" {
		t.Fatalf("%+v", search)
	}
}

func TestGroupRoleMappers(t *testing.T) {
	role := authHttp.GroupRoleFromPB(&userv1.GroupRole{Id: "r1", Name: "Mod", Permissions: nil})
	if role.ID != "r1" || len(role.Permissions) != 0 {
		t.Fatalf("%+v", role)
	}
	roles2 := authHttp.GroupRolesFromPB([]*userv1.GroupRole{{Id: "r1"}})
	if len(roles2) != 1 {
		t.Fatal(roles2)
	}
}
