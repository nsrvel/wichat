package mention_test

import (
	"testing"

	"github.com/wichat/wichat/backend/ms-chat/pkg/mention"
)

func TestPublicUserIDsFromBody_uuid(t *testing.T) {
	body := "hi [@Garam](wichat-mention:9cd306e2-f8ba-4045-82af-7eb2fd0d1b44)"
	got := mention.PublicUserIDsFromBody(body)
	if len(got) != 1 || got[0] != "9cd306e2-f8ba-4045-82af-7eb2fd0d1b44" {
		t.Fatalf("PublicUserIDsFromBody() = %v", got)
	}
}

func TestStripCallerPublicID(t *testing.T) {
	t.Parallel()
	ids := []string{"aaa", "bbb", "aaa"}
	got := mention.StripCallerPublicID(ids, "aaa")
	if len(got) != 1 || got[0] != "bbb" {
		t.Fatalf("StripCallerPublicID() = %v", got)
	}
}

func TestPublicRoleIDsFromBody(t *testing.T) {
	body := "hi [@Mod](wichat-mention-role:aaaaaaaaaaaaaaaaaaaaaaaa)"
	got := mention.PublicRoleIDsFromBody(body)
	if len(got) != 1 || got[0] != "aaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("PublicRoleIDsFromBody() = %v", got)
	}
}

func TestStripTokensForPreview(t *testing.T) {
	body := "hi [@Garam](wichat-mention:aaaaaaaaaaaaaaaaaaaaaaaa) [@Mod](wichat-mention-role:bbbbbbbbbbbbbbbbbbbbbbbb) there"
	got := mention.StripTokensForPreview(body)
	if got != "hi @Garam @Mod there" {
		t.Fatalf("StripTokensForPreview() = %q", got)
	}
}
