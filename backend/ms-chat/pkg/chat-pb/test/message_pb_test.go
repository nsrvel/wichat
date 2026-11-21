package chatPb_test

import (
	"testing"
	"time"

	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	chatPb "github.com/wichat/wichat/backend/ms-chat/pkg/chat-pb"
)

func TestMessageSummary_mapsFields(t *testing.T) {
	edited := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	msg := models.PublicMessage{
		ID: "msg1", AuthorID: "u1", AuthorDisplayName: "Alice", Body: "hi",
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EditedAt:  &edited,
		Mine:      true,
	}
	pb := chatPb.MessageSummary(msg)
	if pb.GetId() != "msg1" || pb.GetAuthorDisplayName() != "Alice" || !pb.GetMine() {
		t.Fatalf("unexpected mapping: %+v", pb)
	}
	if pb.GetEditedAt() == "" {
		t.Fatal("expected edited_at")
	}
}
