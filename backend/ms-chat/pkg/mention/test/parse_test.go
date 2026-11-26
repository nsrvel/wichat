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

func TestStripTokensForPreview(t *testing.T) {
	body := "hi [@Garam](wichat-mention:aaaaaaaaaaaaaaaaaaaaaaaa) there"
	got := mention.StripTokensForPreview(body)
	if got != "hi @Garam there" {
		t.Fatalf("StripTokensForPreview() = %q", got)
	}
}
