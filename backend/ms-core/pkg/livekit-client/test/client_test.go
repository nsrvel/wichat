package livekitclient_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	livekitclient "github.com/wichat/wichat/backend/ms-core/pkg/livekit-client"
	"google.golang.org/protobuf/proto"
)

func TestIssueTokenAndVerifyWebhook(t *testing.T) {
	const apiKey = "devkey"
	const apiSecret = "secretdevkey1234567890abcdef1234567890"
	client := livekitclient.New("ws://localhost:7880", apiKey, apiSecret)

	jwt, exp, err := client.IssueToken(livekitclient.TokenGrants{
		RoomName:        "channel:ch1",
		Identity:        "user-1",
		Name:            "Alice",
		Metadata:        livekitclient.ParticipantMetadataJSON("Alice", "alice", ""),
		TTL:             15 * time.Minute,
		CanPublish:      true,
		CanSubscribe:    true,
		CanPublishAudio: true,
		CanPublishVideo: false,
	})
	if err != nil {
		t.Fatalf("IssueToken() error = %v", err)
	}
	if jwt == "" || exp.Before(time.Now().UTC()) {
		t.Fatalf("invalid token/expiry: len=%d exp=%v", len(jwt), exp)
	}

	// Verify signed webhook body
	rawBody := []byte(`{"id":"ev-1","event":"room_started","room":{"name":"channel:ch1"}}`)
	sum := sha256.Sum256(rawBody)
	sha := base64.StdEncoding.EncodeToString(sum[:])
	at := auth.NewAccessToken(apiKey, apiSecret).SetSha256(sha).SetValidFor(time.Minute)
	authHeader, err := at.ToJWT()
	if err != nil {
		t.Fatalf("sign webhook: %v", err)
	}

	ev, err := client.VerifyWebhook(authHeader, rawBody)
	if err != nil {
		t.Fatalf("VerifyWebhook() error = %v", err)
	}
	if ev.GetEvent() != "room_started" || ev.GetRoom().GetName() != "channel:ch1" {
		t.Fatalf("unexpected webhook event: %+v", ev)
	}

	// Forged signature must fail
	if _, err := client.VerifyWebhook("invalid.jwt.token", rawBody); err == nil {
		t.Fatal("expected error for forged webhook")
	}
}

func TestTwirpRoomAndEgressOperations(t *testing.T) {
	var seenPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPaths = append(seenPaths, r.URL.Path)
		w.Header().Set("Content-Type", "application/protobuf")
		var msg proto.Message
		switch {
		case strings.HasSuffix(r.URL.Path, "UpdateParticipant"):
			msg = &livekit.ParticipantInfo{Identity: "u1"}
		case strings.HasSuffix(r.URL.Path, "GetParticipant"):
			msg = &livekit.ParticipantInfo{
				Identity: "u1",
				Tracks: []*livekit.TrackInfo{
					{Sid: "tr-a", Type: livekit.TrackType_AUDIO},
					{Sid: "tr-v", Type: livekit.TrackType_VIDEO},
				},
			}
		case strings.HasSuffix(r.URL.Path, "MutePublishedTrack"):
			msg = &livekit.MuteRoomTrackResponse{}
		case strings.HasSuffix(r.URL.Path, "RemoveParticipant"):
			msg = &livekit.RemoveParticipantResponse{}
		case strings.HasSuffix(r.URL.Path, "ListParticipants"):
			msg = &livekit.ListParticipantsResponse{Participants: []*livekit.ParticipantInfo{{Identity: "u1"}}}
		case strings.HasSuffix(r.URL.Path, "StartRoomCompositeEgress"):
			msg = &livekit.EgressInfo{EgressId: "eg-1"}
		case strings.HasSuffix(r.URL.Path, "StopEgress"):
			msg = &livekit.EgressInfo{EgressId: "eg-1"}
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		raw, _ := proto.Marshal(msg)
		_, _ = w.Write(raw)
	}))
	defer srv.Close()

	client := livekitclient.New(srv.URL, "devkey", "secretdevkey1234567890abcdef1234567890")
	ctx := context.Background()

	if err := client.UpdateParticipantPermissions(ctx, livekitclient.ParticipantModeration{
		RoomName: "room-1", Identity: "u1", CanPublish: true, CanSubscribe: true,
	}); err != nil {
		t.Fatalf("UpdateParticipantPermissions() error = %v", err)
	}
	if err := client.MuteAudioTracks(ctx, "room-1", "u1", true); err != nil {
		t.Fatalf("MuteAudioTracks() error = %v", err)
	}
	if err := client.MuteVideoTracks(ctx, "room-1", "u1", true); err != nil {
		t.Fatalf("MuteVideoTracks() error = %v", err)
	}
	if err := client.RemoveParticipant(ctx, "room-1", "u1"); err != nil {
		t.Fatalf("RemoveParticipant() error = %v", err)
	}
	parts, err := client.ListParticipants(ctx, "room-1")
	if err != nil || len(parts) != 1 {
		t.Fatalf("ListParticipants() = (%+v, %v)", parts, err)
	}
	egID, err := client.StartRoomCompositeEgress(ctx, "room-1", livekitclient.EgressS3Config{
		Endpoint: "http://minio:9000", Bucket: "wichat", Key: "k", Secret: "s", Filepath: "rec.mp4",
	})
	if err != nil || egID != "eg-1" {
		t.Fatalf("StartRoomCompositeEgress() = (%q, %v)", egID, err)
	}
	if err := client.StopEgress(ctx, "eg-1"); err != nil {
		t.Fatalf("StopEgress() error = %v", err)
	}
	if len(seenPaths) < 8 {
		t.Fatalf("expected >=8 Twirp calls, got %d: %+v", len(seenPaths), seenPaths)
	}
}

func TestDisabledAndEdgeCases(t *testing.T) {
	ctx := context.Background()
	disabled := livekitclient.New("", "", "")
	if _, _, err := disabled.IssueToken(livekitclient.TokenGrants{RoomName: "r", Identity: "u"}); err == nil {
		t.Fatal("expected error when credentials empty")
	}
	if _, err := disabled.VerifyWebhook("h", []byte("{}")); err == nil {
		t.Fatal("expected error when webhook credentials empty")
	}
	if err := disabled.UpdateParticipantPermissions(ctx, livekitclient.ParticipantModeration{}); err != nil {
		t.Fatalf("UpdateParticipantPermissions disabled err = %v", err)
	}
	if err := disabled.MuteAudioTracks(ctx, "r", "u", true); err != nil {
		t.Fatalf("MuteAudioTracks disabled err = %v", err)
	}
	if err := disabled.RemoveParticipant(ctx, "r", "u"); err != nil {
		t.Fatalf("RemoveParticipant disabled err = %v", err)
	}
	if parts, err := disabled.ListParticipants(ctx, "r"); err != nil || parts != nil {
		t.Fatalf("ListParticipants disabled = (%+v, %v)", parts, err)
	}
	if _, err := disabled.StartRoomCompositeEgress(ctx, "r", livekitclient.EgressS3Config{}); err == nil {
		t.Fatal("expected error when egress client disabled")
	}
	if err := disabled.StopEgress(ctx, "eg-1"); err == nil {
		t.Fatal("expected error when stop egress disabled")
	}

	wssClient := livekitclient.New("wss://lk.example.com", "k", "secretdevkey1234567890abcdef1234567890")
	if tok, _, err := wssClient.IssueToken(livekitclient.TokenGrants{
		RoomName: "r", Identity: "u", CanPublish: true, CanPublishAudio: true, CanPublishVideo: true,
	}); err != nil || tok == "" {
		t.Fatalf("wss IssueToken = (%q, %v)", tok, err)
	}

	errSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer errSrv.Close()
	errClient := livekitclient.New(errSrv.URL, "k", "secretdevkey1234567890abcdef1234567890")
	if _, err := errClient.ListParticipants(ctx, "r"); err == nil {
		t.Fatal("expected error on 500 ListParticipants")
	}
	if _, err := errClient.StartRoomCompositeEgress(ctx, "r", livekitclient.EgressS3Config{}); err == nil {
		t.Fatal("expected error on 500 StartRoomCompositeEgress")
	}
}
