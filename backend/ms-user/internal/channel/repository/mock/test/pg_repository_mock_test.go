package mock_test

import (
	"context"
	"testing"
	"time"

	channelmock "github.com/wichat/wichat/backend/ms-user/internal/channel/repository/mock"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
)

func TestChannelMockSeedAndLookup(t *testing.T) {
	st := usermock.NewState()
	repo := channelmock.NewPgRepositoryMock(st)
	ctx := context.Background()
	now := time.Now().UTC()

	if err := repo.SeedDefaultChannels(ctx, "g1", "u1", now); err != nil {
		t.Fatal(err)
	}
	channels, err := repo.ListByGroupID(ctx, "g1")
	if err != nil || len(channels) != 1 {
		t.Fatalf("%+v %v", channels, err)
	}

	ch, err := repo.FindByPublicIDInGroup(ctx, "g1", channels[0].PublicID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindByID(ctx, ch.ID); err != nil {
		t.Fatal(err)
	}
	_, err = repo.FindByPublicIDInGroup(ctx, "g1", ch.PublicID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = repo.FindByPublicIDInGroup(ctx, "g1", "ffffffffffffffffffffffff")
	if err == nil {
		t.Fatal("not found")
	}
	if err := repo.SeedDefaultChannels(ctx, "g1", "u1", now); err != nil {
		t.Fatal(err)
	}
}
