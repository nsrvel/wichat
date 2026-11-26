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

func TestChannelMockGroupDm(t *testing.T) {
	st := usermock.NewState()
	repo := channelmock.NewPgRepositoryMock(st)
	ctx := context.Background()
	now := time.Now().UTC()

	ch, err := repo.CreateGroupDmChannel(ctx, "g1", "Peer", "u1", "u2", "u1", now)
	if err != nil {
		t.Fatal(err)
	}

	found, err := repo.FindGroupDmChannelBetweenUsers(ctx, "g1", "u1", "u2")
	if err != nil || found.ID != ch.ID {
		t.Fatalf("FindGroupDmChannelBetweenUsers: %+v err=%v", found, err)
	}

	visible, err := repo.DmChannelIDSetForUser(ctx, "g1", "u1")
	if err != nil || len(visible) != 1 {
		t.Fatalf("DmChannelIDSetForUser: %+v err=%v", visible, err)
	}
}
