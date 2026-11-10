package invite_test

import (
	"testing"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/pkg/invite"
)

func TestMemberCursorRoundTrip(t *testing.T) {
	t.Parallel()

	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	in := &models.GroupMemberCursor{CreatedAt: ts, UserID: "u1"}
	enc := invite.EncodeMemberCursor(in)
	out, err := invite.DecodeMemberCursor(enc)
	if err != nil || out == nil || !out.CreatedAt.Equal(ts) || out.UserID != "u1" {
		t.Fatalf("round trip: %+v %v", out, err)
	}
}

func TestDecodeMemberCursorEmpty(t *testing.T) {
	t.Parallel()

	out, err := invite.DecodeMemberCursor("   ")
	if err != nil || out != nil {
		t.Fatalf("empty: %+v %v", out, err)
	}
}

func TestDecodeMemberCursorInvalid(t *testing.T) {
	t.Parallel()

	if _, err := invite.DecodeMemberCursor("@@@"); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestEncodeMemberCursorNil(t *testing.T) {
	t.Parallel()

	if invite.EncodeMemberCursor(nil) != "" {
		t.Fatal("nil encodes empty")
	}
}
