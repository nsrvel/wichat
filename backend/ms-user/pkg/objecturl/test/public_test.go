package objecturl_test

import (
	"testing"

	"github.com/wichat/wichat/backend/ms-user/pkg/objecturl"
)

func TestPublic(t *testing.T) {
	t.Parallel()

	if got := objecturl.Public("", "https://cdn/x"); got != "" {
		t.Fatalf("empty %q", got)
	}
	if got := objecturl.Public("https://cdn/x/a.png", "https://other"); got != "https://cdn/x/a.png" {
		t.Fatalf("absolute %q", got)
	}
	if got := objecturl.Public("/groups/g1/icon.png", "https://cdn/bucket"); got != "https://cdn/bucket/groups/g1/icon.png" {
		t.Fatalf("join %q", got)
	}
	if got := objecturl.Public("groups/g1/icon.png", ""); got != "/groups/g1/icon.png" {
		t.Fatalf("relative key %q", got)
	}
}
