package mock_test

import (
	"context"
	"testing"
	"time"

	invitemock "github.com/wichat/wichat/backend/ms-user/internal/invite/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
)

func TestInviteMockLifecycle(t *testing.T) {
	st := usermock.NewState()
	repo := invitemock.NewPgRepositoryMock(st)
	ctx := context.Background()
	now := time.Now().UTC()

	row := models.GroupInviteRow{ID: "i1", Token: "abcdefgh", GroupID: "g1", Kind: constants.GroupInviteKindShareLink, ExpiresAt: now.Add(time.Hour), CreatedAt: now}
	if err := repo.InsertInvite(ctx, row); err != nil {
		t.Fatal(err)
	}
	if err := repo.InsertInvite(ctx, row); err == nil {
		t.Fatal("conflict")
	}

	got, err := repo.FindByToken(ctx, "abcdefgh")
	if err != nil || got.ID != "i1" {
		t.Fatalf("%+v %v", got, err)
	}

	open, err := repo.FindOpenShareLink(ctx, "g1", nil, now)
	if err != nil || open.Token != "abcdefgh" {
		t.Fatalf("%+v %v", open, err)
	}

	if err := repo.MarkRedeemed(ctx, "i1", "u1", now); err != nil {
		t.Fatal(err)
	}
	if err := repo.MarkRedeemed(ctx, "i1", "u1", now); err == nil {
		t.Fatal("already redeemed")
	}
}

func TestFindOpenShareLinkPrefersGroupWideLink(t *testing.T) {
	st := usermock.NewState()
	now := time.Now().UTC()
	ch := "ch-1"
	st.Invites = []models.GroupInviteRow{
		{
			ID: "i1", Token: "12345678", GroupID: "g1", Kind: constants.GroupInviteKindShareLink,
			ChannelID: &ch, ExpiresAt: now.Add(time.Hour), CreatedAt: now,
		},
		{
			ID: "i2", Token: "87654321", GroupID: "g1", Kind: constants.GroupInviteKindShareLink,
			ExpiresAt: now.Add(time.Hour), CreatedAt: now.Add(time.Minute),
		},
	}
	repo := invitemock.NewPgRepositoryMock(st)
	ctx := context.Background()

	row, err := repo.FindOpenShareLink(ctx, "g1", nil, now)
	if err != nil || row.ID != "i2" {
		t.Fatalf("%+v %v", row, err)
	}

	row, err = repo.FindOpenShareLink(ctx, "g1", &ch, now)
	if err != nil || row.ChannelID == nil || *row.ChannelID != ch {
		t.Fatalf("%+v %v", row, err)
	}
}

func TestMarkRedeemedNotFound(t *testing.T) {
	st := usermock.NewState()
	repo := invitemock.NewPgRepositoryMock(st)
	if err := repo.MarkRedeemed(context.Background(), "nope", "u1", time.Now().UTC()); err == nil {
		t.Fatal("expected not found")
	}
}
