package voiceHttp_test

import (
	"testing"

	voiceHttp "github.com/wichat/wichat/backend/api-gateway/pkg/voice-http"
	corev1 "github.com/wichat/wichat/backend/wi-shared/models/gen/core/v1"
)

func TestOccupantsAndRecordingFromPB(t *testing.T) {
	if got := voiceHttp.OccupantFromPB(nil); got.UserID != "" {
		t.Fatalf("OccupantFromPB(nil) = %+v, want empty", got)
	}
	if got := voiceHttp.RecordingFromPB(nil); got.ID != "" {
		t.Fatalf("RecordingFromPB(nil) = %+v, want empty", got)
	}

	occs := voiceHttp.OccupantsFromPB([]*corev1.VoiceOccupant{{
		UserId:           "u1",
		DisplayName:      "Alice",
		Username:         "alice",
		IsMuted:          true,
		IsServerDeafened: true,
	}})
	if len(occs) != 1 || occs[0].UserID != "u1" || !occs[0].IsMuted || !occs[0].IsServerDeafened {
		t.Fatalf("OccupantsFromPB = %+v", occs)
	}

	rec := voiceHttp.RecordingFromPB(&corev1.CallRecordingSummary{
		Id:        "rec-1",
		SessionId: "sess-1",
		ChannelId: "ch-1",
		EgressId:  "eg-1",
		Status:    "completed",
		FileKey:   "recordings/a.mp4",
	})
	if rec.ID != "rec-1" || rec.Status != "completed" || rec.FileKey != "recordings/a.mp4" {
		t.Fatalf("RecordingFromPB = %+v", rec)
	}
}
