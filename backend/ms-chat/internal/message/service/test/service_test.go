package service_test

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/ms-chat/internal/message/repository/mock"
	"github.com/wichat/wichat/backend/ms-chat/internal/message/service"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	outboxmock "github.com/wichat/wichat/backend/ms-chat/internal/outbox/repository/mock"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/channel"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

const testGroupPublicID = "g1"

func testAccessToken(userID string) string {
	if userID == "u2" {
		return "tok-u2"
	}
	return "tok-u1"
}

type fakeChannel struct {
	channel.Client
	perms           []string
	channelType     string
	slowmodeSeconds int32
}

func (f fakeChannel) ListGroupChannels(context.Context, *userv1.ListGroupChannelsRequest) (*userv1.ListGroupChannelsResponse, error) {
	return nil, nil
}

func (f fakeChannel) AssertGroupChannelAccess(_ context.Context, req *userv1.AssertGroupChannelAccessRequest) (*userv1.AssertGroupChannelAccessResponse, error) {
	channelUUID := strings.TrimSpace(req.GetChannelId())
	if channelUUID == "" {
		channelUUID = "channel-uuid"
	}
	perms := f.perms
	if len(perms) == 0 {
		perms = []string{
			"administrator",
			"view_channels",
			"read_message_history",
			"send_messages",
			"manage_channels",
			"add_reactions",
			"manage_messages",
			"pin_messages",
		}
	}
	userUUID := "u1"
	switch req.GetAccessToken() {
	case "tok-u2":
		userUUID = "u2"
	}
	return &userv1.AssertGroupChannelAccessResponse{
		ChannelUuid:     channelUUID,
		UserUuid:        userUUID,
		Permissions:     perms,
		ChannelType:     f.channelType,
		SlowmodeSeconds: f.slowmodeSeconds,
	}, nil
}

func newTestService(st *mock.State, ch fakeChannel) service.Service {
	return service.NewService(
		mock.NewPgRepositoryMock(st),
		nil,
		nil,
		nil,
		ch,
		nil,
		outboxmock.NewPgRepositoryMock(outboxmock.NewState()),
		nil,
	)
}

func TestCreateChannelMessageRejectsEmptyBody(t *testing.T) {
	st := mock.NewState()
	svc := newTestService(st, fakeChannel{})
	_, err := svc.CreateChannelMessage(context.Background(), &models.CreateChannelMessageInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "channel-uuid",
		AccessToken:     testAccessToken("user-uuid"),
		Body:            "   ",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	ex, ok := exception.As(err)
	if !ok || ex.Kind != exception.KindInvalidInput {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func TestCreateChannelMessageRejectsControlChars(t *testing.T) {
	st := mock.NewState()
	svc := newTestService(st, fakeChannel{})
	_, err := svc.CreateChannelMessage(context.Background(), &models.CreateChannelMessageInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "channel-uuid",
		AccessToken:     testAccessToken("user-uuid"),
		Body:            "hello\x00world",
	})
	if err == nil {
		t.Fatal("expected error for null byte")
	}
}

func TestCreateChannelMessageEnforcesPermissionsBroadcastAndSlowmode(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := mock.NewState()
	st.Authors["u1"] = "Alice"
	st.Authors["u2"] = "Bob"

	// 1. Denies when send_messages is missing.
	svcNoSend := newTestService(st, fakeChannel{perms: []string{"view_channels", "read_message_history"}})
	_, err := svcNoSend.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		GroupPublicID:   "g1",
		ChannelPublicID: "ch-1",
		AccessToken:     "tok",
		Body:            "blocked",
	})
	if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindPermissionDenied {
		t.Fatalf("expected permission denied without send_messages, got %v", err)
	}

	// 2. Broadcast channel denies regular member, allows manage_channels.
	svcBroadcast := newTestService(st, fakeChannel{
		channelType: "broadcast",
		perms:       []string{"view_channels", "send_messages"},
	})
	_, err = svcBroadcast.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		GroupPublicID:   "g1",
		ChannelPublicID: "ch-b",
		AccessToken:     "tok",
		Body:            "nope",
	})
	if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindPermissionDenied {
		t.Fatalf("expected permission denied in broadcast for regular member, got %v", err)
	}
	svcBroadcastAdmin := newTestService(st, fakeChannel{
		channelType: "broadcast",
		perms:       []string{"view_channels", "send_messages", "manage_channels"},
	})
	if _, err := svcBroadcastAdmin.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		GroupPublicID:   "g1",
		ChannelPublicID: "ch-b",
		AccessToken:     "tok",
		Body:            "announcement",
	}); err != nil {
		t.Fatalf("expected broadcast post allowed with manage_channels, got %v", err)
	}

	// 3. Slowmode blocks rapid second message for u1, allows u2 with bypass_slowmode.
	svcSlow := newTestService(st, fakeChannel{
		slowmodeSeconds: 30,
		perms:           []string{"view_channels", "send_messages", "read_message_history"},
	})
	if _, err := svcSlow.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		GroupPublicID:   "g1",
		ChannelPublicID: "ch-slow",
		AccessToken:     "tok",
		CallerUserID:    "u1",
		Body:            "first",
	}); err != nil {
		t.Fatalf("first message failed: %v", err)
	}
	_, err = svcSlow.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		GroupPublicID:   "g1",
		ChannelPublicID: "ch-slow",
		AccessToken:     "tok",
		CallerUserID:    "u1",
		Body:            "second",
	})
	if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindRateLimited {
		t.Fatalf("expected rate limited on slowmode, got %v", err)
	}
	svcSlowBypass := newTestService(st, fakeChannel{
		slowmodeSeconds: 30,
		perms:           []string{"view_channels", "send_messages", "bypass_slowmode"},
	})
	if _, err := svcSlowBypass.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		GroupPublicID:   "g1",
		ChannelPublicID: "ch-slow",
		AccessToken:     "tok",
		CallerUserID:    "u2",
		Body:            "bypassed",
	}); err != nil {
		t.Fatalf("expected bypass_slowmode to succeed, got %v", err)
	}

	// 4. ListChannelMessages returns empty when read_message_history is missing.
	svcList := newTestService(st, fakeChannel{perms: []string{"view_channels"}})
	msgs, hasMore, err := svcList.ListChannelMessages(ctx, &models.ListChannelMessagesInput{
		GroupPublicID:   "g1",
		ChannelPublicID: "ch-slow",
		AccessToken:     "tok",
	})
	if err != nil || hasMore || len(msgs) != 0 {
		t.Fatalf("expected empty list without read_message_history, got %d hasMore=%v err=%v", len(msgs), hasMore, err)
	}
}

func TestMessageReplyEditDeleteReactionsAndPins(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := mock.NewState()
	st.Authors["u1"] = "Alice"
	st.Authors["u2"] = "Bob"
	svc := newTestService(st, fakeChannel{})

	// Create root message & reply
	root, err := svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-1",
		AccessToken:     testAccessToken("u1"),
		Body:            "root message",
	})
	if err != nil {
		t.Fatal(err)
	}
	reply, err := svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-1",
		AccessToken:     testAccessToken("u2"),
		Body:            "reply body",
		ReplyToPublicID: root.ID,
	})
	if err != nil || reply.ReplyTo == nil || reply.ReplyTo.ID != root.ID {
		t.Fatalf("expected reply preview to %s, got %+v err=%v", root.ID, reply.ReplyTo, err)
	}

	// Edit: non-author denied, author succeeds
	if _, err := svc.UpdateChannelMessage(ctx, &models.UpdateChannelMessageInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-1",
		AccessToken:     testAccessToken("u2"),
		MessagePublicID: root.ID,
		Body:            "hacked",
	}); err == nil {
		t.Fatal("expected permission denied when editing another user's message")
	}
	edited, err := svc.UpdateChannelMessage(ctx, &models.UpdateChannelMessageInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-1",
		AccessToken:     testAccessToken("u1"),
		MessagePublicID: root.ID,
		Body:            "edited root",
	})
	if err != nil || edited.Body != "edited root" || edited.EditedAt == nil {
		t.Fatalf("unexpected edited message: %+v err=%v", edited, err)
	}

	// Reactions: denied without add_reactions, toggles on/off with add_reactions
	svcNoReact := newTestService(st, fakeChannel{perms: []string{"view_channels", "read_message_history"}})
	if _, err := svcNoReact.ToggleMessageReaction(ctx, &models.ToggleMessageReactionInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-1",
		AccessToken:     testAccessToken("u2"),
		MessagePublicID: root.ID,
		Emoji:           "👍",
	}); err == nil {
		t.Fatal("expected permission denied without add_reactions")
	}
	r1, err := svc.ToggleMessageReaction(ctx, &models.ToggleMessageReactionInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-1",
		AccessToken:     testAccessToken("u2"),
		MessagePublicID: root.ID,
		Emoji:           "👍",
	})
	if err != nil || len(r1) != 1 || r1[0].Count != 1 {
		t.Fatalf("expected 1 reaction, got %+v err=%v", r1, err)
	}
	r2, err := svc.ToggleMessageReaction(ctx, &models.ToggleMessageReactionInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-1",
		AccessToken:     testAccessToken("u2"),
		MessagePublicID: root.ID,
		Emoji:           "👍",
	})
	if err != nil || len(r2) != 0 {
		t.Fatalf("expected reaction removed on toggle off, got %+v err=%v", r2, err)
	}

	// Pin/Unpin: requires pin_messages
	svcNoPin := newTestService(st, fakeChannel{perms: []string{"view_channels", "read_message_history", "add_reactions"}})
	if _, err := svcNoPin.SetChannelMessagePinned(ctx, &models.SetChannelMessagePinnedInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-1",
		AccessToken:     testAccessToken("u2"),
		MessagePublicID: root.ID,
		Pinned:          true,
	}); err == nil {
		t.Fatal("expected permission denied without pin_messages")
	}
	pinned, err := svc.SetChannelMessagePinned(ctx, &models.SetChannelMessagePinnedInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-1",
		AccessToken:     testAccessToken("u1"),
		MessagePublicID: root.ID,
		Pinned:          true,
	})
	if err != nil || !pinned.IsPinned {
		t.Fatalf("expected pinned, got %+v err=%v", pinned, err)
	}
	pins, err := svc.ListPinnedChannelMessages(ctx, &models.ListPinnedChannelMessagesInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-1",
		AccessToken:     testAccessToken("u2"),
	})
	if err != nil || len(pins) != 1 || pins[0].ID != root.ID {
		t.Fatalf("expected 1 pinned message, got %+v err=%v", pins, err)
	}

	// Delete: non-author denied without manage_messages, allowed with manage_messages
	svcNoManage := newTestService(st, fakeChannel{perms: []string{"view_channels", "read_message_history"}})
	if err := svcNoManage.DeleteChannelMessage(ctx, &models.DeleteChannelMessageInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-1",
		AccessToken:     testAccessToken("u2"),
		MessagePublicID: root.ID,
	}); err == nil {
		t.Fatal("expected permission denied deleting other's message")
	}
	svcCanManage := newTestService(st, fakeChannel{perms: []string{"view_channels", "read_message_history", "manage_messages"}})
	if err := svcCanManage.DeleteChannelMessage(ctx, &models.DeleteChannelMessageInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-1",
		AccessToken:     testAccessToken("u2"),
		MessagePublicID: root.ID,
	}); err != nil {
		t.Fatalf("expected manage_messages delete to succeed: %v", err)
	}
}

type fakeMinioRepo struct {
	lastChannelID   string
	lastFilename    string
	lastContentType string
	lastSize        int64
	err             error
}

func (f *fakeMinioRepo) PutChannelAttachment(_ context.Context, channelID, attachmentID, filename, contentType string, _ io.Reader, size int64) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	f.lastChannelID = channelID
	f.lastFilename = filename
	f.lastContentType = contentType
	f.lastSize = size
	return "http://localhost:9000/wichat-public/channels/" + channelID + "/attachments/" + attachmentID + "/" + filename, nil
}

func TestUploadChannelAttachment(t *testing.T) {
	ctx := context.Background()
	st := mock.NewState()
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D}
	pdfHeader := []byte("%PDF-1.7\nhello")

	t.Run("uploads valid PNG and sanitizes filename", func(t *testing.T) {
		repo := &fakeMinioRepo{}
		svc := service.NewService(mock.NewPgRepositoryMock(st), nil, repo, nil, fakeChannel{}, nil, outboxmock.NewPgRepositoryMock(outboxmock.NewState()), nil)
		att, err := svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			GroupPublicID:     testGroupPublicID,
			ChannelPublicID:   "ch-1",
			AccessToken:       testAccessToken("u1"),
			CallerUserID:      "u1",
			CallerPermissions: []string{"send_messages", "attach_files"},
			Filename:          "../../my<pic>.png",
			ContentType:       "image/png",
			Data:              pngHeader,
		})
		if err != nil {
			t.Fatalf("UploadChannelAttachment() error = %v", err)
		}
		if att.Filename != "my_pic_.png" {
			t.Fatalf("Filename = %q, want my_pic_.png", att.Filename)
		}
		if att.ContentType != "image/png" {
			t.Fatalf("ContentType = %q, want image/png", att.ContentType)
		}
		if att.SizeBytes != int64(len(pngHeader)) {
			t.Fatalf("SizeBytes = %d, want %d", att.SizeBytes, len(pngHeader))
		}
		if att.ID == "" || att.URL == "" {
			t.Fatalf("expected non-empty ID and URL, got %+v", att)
		}
	})

	t.Run("uploads PDF and markdown", func(t *testing.T) {
		repo := &fakeMinioRepo{}
		svc := service.NewService(mock.NewPgRepositoryMock(st), nil, repo, nil, fakeChannel{}, nil, outboxmock.NewPgRepositoryMock(outboxmock.NewState()), nil)

		pdfAtt, err := svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			GroupPublicID:     testGroupPublicID,
			ChannelPublicID:   "ch-1",
			AccessToken:       testAccessToken("u1"),
			CallerUserID:      "u1",
			CallerPermissions: []string{"administrator"},
			Filename:          "spec.pdf",
			ContentType:       "application/pdf",
			Data:              pdfHeader,
		})
		if err != nil || pdfAtt.ContentType != "application/pdf" {
			t.Fatalf("PDF upload failed: %+v err=%v", pdfAtt, err)
		}

		mdAtt, err := svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			GroupPublicID:     testGroupPublicID,
			ChannelPublicID:   "ch-1",
			AccessToken:       testAccessToken("u1"),
			CallerUserID:      "u1",
			CallerPermissions: []string{"administrator"},
			Filename:          "notes.md",
			ContentType:       "text/markdown",
			Data:              []byte("# Heading\nhello"),
		})
		if err != nil || mdAtt.ContentType != "text/markdown" {
			t.Fatalf("Markdown upload failed: %+v err=%v", mdAtt, err)
		}
	})

	t.Run("rejects HTML/SVG/executable and oversized files", func(t *testing.T) {
		repo := &fakeMinioRepo{}
		svc := service.NewService(mock.NewPgRepositoryMock(st), nil, repo, nil, fakeChannel{}, nil, outboxmock.NewPgRepositoryMock(outboxmock.NewState()), nil)

		_, err := svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			GroupPublicID:   testGroupPublicID,
			ChannelPublicID: "ch-1",
			AccessToken:     testAccessToken("u1"),
			CallerUserID:    "u1",
			Filename:        "xss.html",
			ContentType:     "text/html",
			Data:            []byte("<html><script>alert(1)</script></html>"),
		})
		if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindInvalidInput {
			t.Fatalf("expected InvalidInput for HTML, got %v", err)
		}

		_, err = svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			GroupPublicID:   testGroupPublicID,
			ChannelPublicID: "ch-1",
			AccessToken:     testAccessToken("u1"),
			CallerUserID:    "u1",
			Filename:        "icon.svg",
			ContentType:     "image/svg+xml",
			Data:            []byte("<svg></svg>"),
		})
		if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindInvalidInput {
			t.Fatalf("expected InvalidInput for SVG, got %v", err)
		}

		_, err = svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			GroupPublicID:   testGroupPublicID,
			ChannelPublicID: "ch-1",
			AccessToken:     testAccessToken("u1"),
			CallerUserID:    "u1",
			Filename:        "empty.png",
			ContentType:     "image/png",
			Data:            nil,
		})
		if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindInvalidInput {
			t.Fatalf("expected InvalidInput for empty file, got %v", err)
		}

		oversized := make([]byte, constants.MaxAttachmentBytes+1)
		_, err = svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			GroupPublicID:   testGroupPublicID,
			ChannelPublicID: "ch-1",
			AccessToken:     testAccessToken("u1"),
			CallerUserID:    "u1",
			Filename:        "big.png",
			ContentType:     "image/png",
			Data:            oversized,
		})
		if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindInvalidInput {
			t.Fatalf("expected InvalidInput for oversized file, got %v", err)
		}
	})

	t.Run("enforces attach_files and broadcast permissions", func(t *testing.T) {
		repo := &fakeMinioRepo{}
		svcNoAttach := service.NewService(mock.NewPgRepositoryMock(st), nil, repo, nil, fakeChannel{perms: []string{"send_messages"}}, nil, outboxmock.NewPgRepositoryMock(outboxmock.NewState()), nil)

		_, err := svcNoAttach.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			GroupPublicID:   testGroupPublicID,
			ChannelPublicID: "ch-1",
			AccessToken:     testAccessToken("u1"),
			Filename:        "pic.png",
			ContentType:     "image/png",
			Data:            pngHeader,
		})
		if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindPermissionDenied {
			t.Fatalf("expected PermissionDenied without attach_files, got %v", err)
		}

		svcBroadcastAttach := service.NewService(mock.NewPgRepositoryMock(st), nil, repo, nil, fakeChannel{
			channelType: "broadcast",
			perms:       []string{"send_messages", "attach_files"},
		}, nil, outboxmock.NewPgRepositoryMock(outboxmock.NewState()), nil)
		_, err = svcBroadcastAttach.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			GroupPublicID:   testGroupPublicID,
			ChannelPublicID: "ch-1",
			AccessToken:     testAccessToken("u1"),
			Filename:        "pic.png",
			ContentType:     "image/png",
			Data:            pngHeader,
		})
		if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindPermissionDenied {
			t.Fatalf("expected PermissionDenied for broadcast channel, got %v", err)
		}
	})
}

func TestListChannelMessagesPaginationAndHasMore(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := mock.NewState()
	st.Authors["u1"] = "Alice"
	svc := newTestService(st, fakeChannel{})

	for _, body := range []string{"m1", "m2", "m3"} {
		if _, err := svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
			GroupPublicID:     testGroupPublicID,
			ChannelPublicID:   "ch-page",
			AccessToken:       testAccessToken("u1"),
			CallerUserID:      "u1",
			CallerPermissions: []string{"view_channels", "send_messages", "read_message_history"},
			Body:              body,
		}); err != nil {
			t.Fatal(err)
		}
	}

	page1, hasMore1, err := svc.ListChannelMessages(ctx, &models.ListChannelMessagesInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-page",
		AccessToken:     testAccessToken("u1"),
		Limit:           2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !hasMore1 || len(page1) != 2 || page1[0].Body != "m2" || page1[1].Body != "m3" {
		t.Fatalf("unexpected page1: %+v hasMore=%v", page1, hasMore1)
	}

	page2, hasMore2, err := svc.ListChannelMessages(ctx, &models.ListChannelMessagesInput{
		GroupPublicID:    testGroupPublicID,
		ChannelPublicID:  "ch-page",
		AccessToken:      testAccessToken("u1"),
		Limit:            2,
		BeforeChannelSeq: page1[0].ChannelSeq,
	})
	if err != nil {
		t.Fatal(err)
	}
	if hasMore2 || len(page2) != 1 || page2[0].Body != "m1" {
		t.Fatalf("unexpected page2: %+v hasMore=%v", page2, hasMore2)
	}
}

func TestUpdateDeleteChannelMessageOutboxWithRepository(t *testing.T) {
	ctx := context.Background()
	st := mock.NewState()
	st.Authors["u1"] = "Alice"
	svc := newTestService(st, fakeChannel{})

	root, err := svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		GroupPublicID:   "grp-1",
		ChannelPublicID: "ch-outbox",
		AccessToken:     testAccessToken("u1"),
		Body:            "original",
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := svc.UpdateChannelMessage(ctx, &models.UpdateChannelMessageInput{
		GroupPublicID:   "grp-1",
		ChannelPublicID: "ch-outbox",
		AccessToken:     testAccessToken("u1"),
		MessagePublicID: root.ID,
		Body:            "edited",
	}); err != nil {
		t.Fatal(err)
	}
	if len(st.Outbox) != 2 || st.Outbox[1].EventType != chatEvents.TypeMessageUpdated {
		t.Fatalf("expected message.updated outbox row, got %+v", st.Outbox)
	}

	st.FailOutbox = true
	if _, err := svc.UpdateChannelMessage(ctx, &models.UpdateChannelMessageInput{
		GroupPublicID:   "grp-1",
		ChannelPublicID: "ch-outbox",
		AccessToken:     testAccessToken("u1"),
		MessagePublicID: root.ID,
		Body:            "should not stick",
	}); err == nil {
		t.Fatal("expected error when outbox fails")
	}
	got, _, err := svc.ListChannelMessages(ctx, &models.ListChannelMessagesInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-outbox",
		AccessToken:     testAccessToken("u1"),
		Limit:           10,
	})
	if err != nil || len(got) != 1 || got[0].Body != "edited" {
		t.Fatalf("expected body unchanged after failed outbox, got %+v err=%v", got, err)
	}

	st.FailOutbox = false
	if err := svc.DeleteChannelMessage(ctx, &models.DeleteChannelMessageInput{
		GroupPublicID:   "grp-1",
		ChannelPublicID: "ch-outbox",
		AccessToken:     testAccessToken("u1"),
		MessagePublicID: root.ID,
	}); err != nil {
		t.Fatal(err)
	}
	if len(st.Outbox) != 3 || st.Outbox[2].EventType != chatEvents.TypeMessageDeleted {
		t.Fatalf("expected message.deleted outbox row, got %+v", st.Outbox)
	}
	got, _, err = svc.ListChannelMessages(ctx, &models.ListChannelMessagesInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-outbox",
		AccessToken:     testAccessToken("u1"),
		Limit:           10,
	})
	if err != nil || len(got) != 0 {
		t.Fatalf("expected message deleted, got %+v err=%v", got, err)
	}
}

func TestCreateChannelMessageClientMsgID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := mock.NewState()
	st.Authors["u1"] = "Alice"
	svc := newTestService(st, fakeChannel{})

	clientID := "550e8400-e29b-41d4-a716-446655440000"
	first, err := svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-client",
		AccessToken:     testAccessToken("u1"),
		Body:            "hello once",
		ClientMsgID:     clientID,
	})
	if err != nil || first.ClientMsgID != clientID {
		t.Fatalf("first create: %+v err=%v", first, err)
	}

	second, err := svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-client",
		AccessToken:     testAccessToken("u1"),
		Body:            "hello once",
		ClientMsgID:     clientID,
	})
	if err != nil || second.ID != first.ID {
		t.Fatalf("idempotent retry: want same id %s got %+v err=%v", first.ID, second, err)
	}

	if _, err := svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-client",
		AccessToken:     testAccessToken("u1"),
		Body:            "other",
		ClientMsgID:     "not-a-uuid",
	}); err == nil {
		t.Fatal("expected invalid client_msg_id")
	}
}

func TestReactionAndPinOutboxSameTransaction(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := mock.NewState()
	st.Authors["u1"] = "Alice"
	st.Authors["u2"] = "Bob"
	svc := newTestService(st, fakeChannel{})

	root, err := svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-tx",
		AccessToken:     testAccessToken("u1"),
		Body:            "pin me",
	})
	if err != nil {
		t.Fatal(err)
	}
	st.Outbox = nil

	if _, err := svc.ToggleMessageReaction(ctx, &models.ToggleMessageReactionInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-tx",
		AccessToken:     testAccessToken("u2"),
		MessagePublicID: root.ID,
		Emoji:           "🔥",
	}); err != nil {
		t.Fatal(err)
	}
	if len(st.Outbox) != 1 || st.Outbox[0].EventType != chatEvents.TypeMessageReactionsUpdated {
		t.Fatalf("reaction outbox: %+v", st.Outbox)
	}

	st.Outbox = nil
	if _, err := svc.SetChannelMessagePinned(ctx, &models.SetChannelMessagePinnedInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-tx",
		AccessToken:     testAccessToken("u1"),
		MessagePublicID: root.ID,
		Pinned:          true,
	}); err != nil {
		t.Fatal(err)
	}
	if len(st.Outbox) != 1 || st.Outbox[0].EventType != chatEvents.TypeMessageUpdated {
		t.Fatalf("pin outbox: %+v", st.Outbox)
	}

	st.FailOutbox = true
	if _, err := svc.ToggleMessageReaction(ctx, &models.ToggleMessageReactionInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-tx",
		AccessToken:     testAccessToken("u2"),
		MessagePublicID: root.ID,
		Emoji:           "🔥",
	}); err == nil {
		t.Fatal("expected error when reaction outbox fails")
	}
}

func TestResolveMessageChannelSeq(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := mock.NewState()
	svc := newTestService(st, fakeChannel{})

	created, err := svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		GroupPublicID:     testGroupPublicID,
		ChannelPublicID:   "ch-seq",
		AccessToken:       testAccessToken("u1"),
		CallerUserID:      "u1",
		CallerPermissions: []string{"view_channels", "send_messages", "read_message_history"},
		Body:              "target",
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := svc.ResolveMessageChannelSeq(ctx, &models.ResolveMessageChannelSeqInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-seq",
		AccessToken:     testAccessToken("u1"),
		MessagePublicID: created.ID,
	})
	if err != nil || got.ChannelSeq <= 0 || got.MessagePublicID != created.ID {
		t.Fatalf("resolve: %+v err=%v", got, err)
	}

	svcNoRead := newTestService(st, fakeChannel{perms: []string{"view_channels"}})
	if _, err := svcNoRead.ResolveMessageChannelSeq(ctx, &models.ResolveMessageChannelSeqInput{
		GroupPublicID:   testGroupPublicID,
		ChannelPublicID: "ch-seq",
		AccessToken:     testAccessToken("u1"),
		MessagePublicID: created.ID,
	}); err == nil {
		t.Fatal("expected permission denied without read_message_history")
	}
}
