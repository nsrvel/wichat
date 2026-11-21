package service_test

import (
	"context"
	"io"
	"testing"

	"github.com/wichat/wichat/backend/ms-chat/internal/message/repository/mock"
	"github.com/wichat/wichat/backend/ms-chat/internal/message/service"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	outboxmock "github.com/wichat/wichat/backend/ms-chat/internal/outbox/repository/mock"
	"github.com/wichat/wichat/backend/ms-chat/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/channel"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

type fakeChannel struct {
	channel.Client
}

func (fakeChannel) ListGroupChannels(context.Context, *userv1.ListGroupChannelsRequest) (*userv1.ListGroupChannelsResponse, error) {
	return nil, nil
}

func (fakeChannel) AssertGroupChannelAccess(context.Context, *userv1.AssertGroupChannelAccessRequest) (*userv1.AssertGroupChannelAccessResponse, error) {
	return &userv1.AssertGroupChannelAccessResponse{
		ChannelUuid: "channel-uuid",
		UserUuid:    "user-uuid",
	}, nil
}

func newTestService(st *mock.State) service.Service {
	return service.NewService(
		mock.NewPgRepositoryMock(st),
		nil,
		nil,
		fakeChannel{},
		outboxmock.NewPgRepositoryMock(outboxmock.NewState()),
		nil,
	)
}

func TestCreateChannelMessageRejectsEmptyBody(t *testing.T) {
	st := mock.NewState()
	svc := newTestService(st)
	_, err := svc.CreateChannelMessage(context.Background(), &models.CreateChannelMessageInput{
		ResolvedChannelID: "channel-uuid",
		CallerUserID:      "user-uuid",
		Body:              "   ",
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
	svc := newTestService(st)
	_, err := svc.CreateChannelMessage(context.Background(), &models.CreateChannelMessageInput{
		ResolvedChannelID: "channel-uuid",
		CallerUserID:      "user-uuid",
		Body:              "hello\x00world",
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
	svc := newTestService(st)

	// 1. Denies when send_messages is missing.
	_, err := svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		ResolvedChannelID: "ch-1",
		CallerUserID:      "u1",
		CallerPermissions: []string{"view_channels", "read_message_history"},
		Body:              "blocked",
	})
	if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindPermissionDenied {
		t.Fatalf("expected permission denied without send_messages, got %v", err)
	}

	// 2. Broadcast channel denies regular member, allows manage_channels.
	_, err = svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		ResolvedChannelID: "ch-b",
		CallerUserID:      "u1",
		CallerPermissions: []string{"view_channels", "send_messages"},
		ChannelType:       "broadcast",
		Body:              "nope",
	})
	if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindPermissionDenied {
		t.Fatalf("expected permission denied in broadcast for regular member, got %v", err)
	}
	if _, err := svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		ResolvedChannelID: "ch-b",
		CallerUserID:      "u1",
		CallerPermissions: []string{"view_channels", "send_messages", "manage_channels"},
		ChannelType:       "broadcast",
		Body:              "announcement",
	}); err != nil {
		t.Fatalf("expected broadcast post allowed with manage_channels, got %v", err)
	}

	// 3. Slowmode blocks rapid second message for u1, allows u2 with bypass_slowmode.
	if _, err := svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		ResolvedChannelID: "ch-slow",
		CallerUserID:      "u1",
		CallerPermissions: []string{"view_channels", "send_messages"},
		SlowmodeSeconds:   30,
		Body:              "first",
	}); err != nil {
		t.Fatalf("first message failed: %v", err)
	}
	_, err = svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		ResolvedChannelID: "ch-slow",
		CallerUserID:      "u1",
		CallerPermissions: []string{"view_channels", "send_messages"},
		SlowmodeSeconds:   30,
		Body:              "second",
	})
	if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindRateLimited {
		t.Fatalf("expected rate limited on slowmode, got %v", err)
	}
	if _, err := svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		ResolvedChannelID: "ch-slow",
		CallerUserID:      "u2",
		CallerPermissions: []string{"view_channels", "send_messages", "bypass_slowmode"},
		SlowmodeSeconds:   30,
		Body:              "bypassed",
	}); err != nil {
		t.Fatalf("expected bypass_slowmode to succeed, got %v", err)
	}

	// 4. ListChannelMessages returns empty when read_message_history is missing.
	msgs, hasMore, err := svc.ListChannelMessages(ctx, &models.ListChannelMessagesInput{
		ResolvedChannelID: "ch-slow",
		CallerUserID:      "u1",
		CallerPermissions: []string{"view_channels"},
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
	svc := newTestService(st)

	// Create root message & reply
	root, err := svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		ResolvedChannelID: "ch-1",
		CallerUserID:      "u1",
		CallerPermissions: []string{"view_channels", "send_messages"},
		Body:              "root message",
	})
	if err != nil {
		t.Fatal(err)
	}
	reply, err := svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
		ResolvedChannelID: "ch-1",
		CallerUserID:      "u2",
		CallerPermissions: []string{"view_channels", "send_messages", "read_message_history"},
		Body:              "reply body",
		ReplyToPublicID:   root.ID,
	})
	if err != nil || reply.ReplyTo == nil || reply.ReplyTo.ID != root.ID {
		t.Fatalf("expected reply preview to %s, got %+v err=%v", root.ID, reply.ReplyTo, err)
	}

	// Edit: non-author denied, author succeeds
	if _, err := svc.UpdateChannelMessage(ctx, &models.UpdateChannelMessageInput{
		ResolvedChannelID: "ch-1",
		CallerUserID:      "u2",
		MessagePublicID:   root.ID,
		Body:              "hacked",
	}); err == nil {
		t.Fatal("expected permission denied when editing another user's message")
	}
	edited, err := svc.UpdateChannelMessage(ctx, &models.UpdateChannelMessageInput{
		ResolvedChannelID: "ch-1",
		CallerUserID:      "u1",
		MessagePublicID:   root.ID,
		Body:              "edited root",
	})
	if err != nil || edited.Body != "edited root" || edited.EditedAt == nil {
		t.Fatalf("unexpected edited message: %+v err=%v", edited, err)
	}

	// Reactions: denied without add_reactions, toggles on/off with add_reactions
	if _, err := svc.ToggleMessageReaction(ctx, &models.ToggleMessageReactionInput{
		ResolvedChannelID: "ch-1",
		CallerUserID:      "u2",
		CallerPermissions: []string{"view_channels"},
		MessagePublicID:   root.ID,
		Emoji:             "👍",
	}); err == nil {
		t.Fatal("expected permission denied without add_reactions")
	}
	r1, err := svc.ToggleMessageReaction(ctx, &models.ToggleMessageReactionInput{
		ResolvedChannelID: "ch-1",
		CallerUserID:      "u2",
		CallerPermissions: []string{"view_channels", "add_reactions"},
		MessagePublicID:   root.ID,
		Emoji:             "👍",
	})
	if err != nil || len(r1) != 1 || r1[0].Count != 1 {
		t.Fatalf("expected 1 reaction, got %+v err=%v", r1, err)
	}
	r2, err := svc.ToggleMessageReaction(ctx, &models.ToggleMessageReactionInput{
		ResolvedChannelID: "ch-1",
		CallerUserID:      "u2",
		CallerPermissions: []string{"view_channels"},
		MessagePublicID:   root.ID,
		Emoji:             "👍",
	})
	if err != nil || len(r2) != 0 {
		t.Fatalf("expected reaction removed on toggle off, got %+v err=%v", r2, err)
	}

	// Pin/Unpin: requires pin_messages
	if _, err := svc.SetChannelMessagePinned(ctx, &models.SetChannelMessagePinnedInput{
		ResolvedChannelID: "ch-1",
		CallerUserID:      "u2",
		CallerPermissions: []string{"view_channels"},
		MessagePublicID:   root.ID,
		Pinned:            true,
	}); err == nil {
		t.Fatal("expected permission denied without pin_messages")
	}
	pinned, err := svc.SetChannelMessagePinned(ctx, &models.SetChannelMessagePinnedInput{
		ResolvedChannelID: "ch-1",
		CallerUserID:      "u1",
		CallerPermissions: []string{"view_channels", "pin_messages"},
		MessagePublicID:   root.ID,
		Pinned:            true,
	})
	if err != nil || !pinned.IsPinned {
		t.Fatalf("expected pinned, got %+v err=%v", pinned, err)
	}
	pins, err := svc.ListPinnedChannelMessages(ctx, &models.ListPinnedChannelMessagesInput{
		ResolvedChannelID: "ch-1",
		CallerUserID:      "u2",
		CallerPermissions: []string{"view_channels", "read_message_history"},
	})
	if err != nil || len(pins) != 1 || pins[0].ID != root.ID {
		t.Fatalf("expected 1 pinned message, got %+v err=%v", pins, err)
	}

	// Delete: non-author denied without manage_messages, allowed with manage_messages
	if err := svc.DeleteChannelMessage(ctx, &models.DeleteChannelMessageInput{
		ResolvedChannelID: "ch-1",
		CallerUserID:      "u2",
		CallerPermissions: []string{"view_channels"},
		MessagePublicID:   root.ID,
	}); err == nil {
		t.Fatal("expected permission denied deleting other's message")
	}
	if err := svc.DeleteChannelMessage(ctx, &models.DeleteChannelMessageInput{
		ResolvedChannelID: "ch-1",
		CallerUserID:      "u2",
		CallerPermissions: []string{"view_channels", "manage_messages"},
		MessagePublicID:   root.ID,
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
		svc := service.NewService(mock.NewPgRepositoryMock(st), repo, nil, fakeChannel{}, outboxmock.NewPgRepositoryMock(outboxmock.NewState()), nil)
		att, err := svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			ResolvedChannelID: "ch-1",
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
		svc := service.NewService(mock.NewPgRepositoryMock(st), repo, nil, fakeChannel{}, outboxmock.NewPgRepositoryMock(outboxmock.NewState()), nil)

		pdfAtt, err := svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			ResolvedChannelID: "ch-1",
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
			ResolvedChannelID: "ch-1",
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
		svc := service.NewService(mock.NewPgRepositoryMock(st), repo, nil, fakeChannel{}, outboxmock.NewPgRepositoryMock(outboxmock.NewState()), nil)

		_, err := svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			ResolvedChannelID: "ch-1",
			CallerUserID:      "u1",
			Filename:          "xss.html",
			ContentType:       "text/html",
			Data:              []byte("<html><script>alert(1)</script></html>"),
		})
		if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindInvalidInput {
			t.Fatalf("expected InvalidInput for HTML, got %v", err)
		}

		_, err = svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			ResolvedChannelID: "ch-1",
			CallerUserID:      "u1",
			Filename:          "icon.svg",
			ContentType:       "image/svg+xml",
			Data:              []byte("<svg></svg>"),
		})
		if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindInvalidInput {
			t.Fatalf("expected InvalidInput for SVG, got %v", err)
		}

		_, err = svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			ResolvedChannelID: "ch-1",
			CallerUserID:      "u1",
			Filename:          "empty.png",
			ContentType:       "image/png",
			Data:              nil,
		})
		if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindInvalidInput {
			t.Fatalf("expected InvalidInput for empty file, got %v", err)
		}

		oversized := make([]byte, constants.MaxAttachmentBytes+1)
		_, err = svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			ResolvedChannelID: "ch-1",
			CallerUserID:      "u1",
			Filename:          "big.png",
			ContentType:       "image/png",
			Data:              oversized,
		})
		if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindInvalidInput {
			t.Fatalf("expected InvalidInput for oversized file, got %v", err)
		}
	})

	t.Run("enforces attach_files and broadcast permissions", func(t *testing.T) {
		repo := &fakeMinioRepo{}
		svc := service.NewService(mock.NewPgRepositoryMock(st), repo, nil, fakeChannel{}, outboxmock.NewPgRepositoryMock(outboxmock.NewState()), nil)

		_, err := svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			ResolvedChannelID: "ch-1",
			CallerUserID:      "u1",
			CallerPermissions: []string{"send_messages"},
			Filename:          "pic.png",
			ContentType:       "image/png",
			Data:              pngHeader,
		})
		if ex, ok := exception.As(err); !ok || ex.Kind != exception.KindPermissionDenied {
			t.Fatalf("expected PermissionDenied without attach_files, got %v", err)
		}

		_, err = svc.UploadChannelAttachment(ctx, &models.UploadChannelAttachmentInput{
			ResolvedChannelID: "ch-1",
			CallerUserID:      "u1",
			CallerPermissions: []string{"send_messages", "attach_files"},
			ChannelType:       "broadcast",
			Filename:          "pic.png",
			ContentType:       "image/png",
			Data:              pngHeader,
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
	svc := newTestService(st)

	for _, body := range []string{"m1", "m2", "m3"} {
		if _, err := svc.CreateChannelMessage(ctx, &models.CreateChannelMessageInput{
			ResolvedChannelID: "ch-page",
			CallerUserID:      "u1",
			CallerPermissions: []string{"view_channels", "send_messages", "read_message_history"},
			Body:              body,
		}); err != nil {
			t.Fatal(err)
		}
	}

	page1, hasMore1, err := svc.ListChannelMessages(ctx, &models.ListChannelMessagesInput{
		ResolvedChannelID: "ch-page",
		CallerUserID:      "u1",
		CallerPermissions: []string{"view_channels", "read_message_history"},
		Limit:             2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !hasMore1 || len(page1) != 2 || page1[0].Body != "m2" || page1[1].Body != "m3" {
		t.Fatalf("unexpected page1: %+v hasMore=%v", page1, hasMore1)
	}

	page2, hasMore2, err := svc.ListChannelMessages(ctx, &models.ListChannelMessagesInput{
		ResolvedChannelID: "ch-page",
		CallerUserID:      "u1",
		CallerPermissions: []string{"view_channels", "read_message_history"},
		Limit:             2,
		BeforeCreatedAt:   &page1[0].CreatedAt,
		BeforePublicID:    page1[0].ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if hasMore2 || len(page2) != 1 || page2[0].Body != "m1" {
		t.Fatalf("unexpected page2: %+v hasMore=%v", page2, hasMore2)
	}
}
