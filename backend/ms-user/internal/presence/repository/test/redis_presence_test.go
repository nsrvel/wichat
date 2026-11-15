package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	"github.com/wichat/wichat/backend/ms-user/internal/presence/repository"
)

func newTestRepo(t *testing.T) (repository.RedisRepository, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return repository.NewRedisRepository(rdb), mr
}

func TestHeartbeatAndMultiTabCleanup(t *testing.T) {
	repo, mr := newTestRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.SetConnHeartbeat(ctx, "u1", "c1", models.PresenceActivityActive, 60*time.Second))
	require.NoError(t, repo.SetConnHeartbeat(ctx, "u1", "c2", models.PresenceActivityIdle, 60*time.Second))

	activities, err := repo.ListActiveConnActivities(ctx, "u1")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{models.PresenceActivityActive, models.PresenceActivityIdle}, activities)

	require.NoError(t, repo.DeleteConn(ctx, "u1", "c1"))
	activities, err = repo.ListActiveConnActivities(ctx, "u1")
	require.NoError(t, err)
	require.Equal(t, []string{models.PresenceActivityIdle}, activities)

	mr.FastForward(61 * time.Second)
	activities, err = repo.ListActiveConnActivities(ctx, "u1")
	require.NoError(t, err)
	require.Empty(t, activities)
}

func TestManualAndBatchEffectivePresence(t *testing.T) {
	repo, _ := newTestRepo(t)
	ctx := context.Background()

	state, err := repo.GetManualStatus(ctx, "u1")
	require.NoError(t, err)
	require.Equal(t, models.PresenceStatusOnline, state.Status)
	require.Nil(t, state.ExpiresAt)

	require.NoError(t, repo.SetManualStatus(ctx, "u1", models.PresenceStatusDND, 30*time.Minute))
	state, err = repo.GetManualStatus(ctx, "u1")
	require.NoError(t, err)
	require.Equal(t, models.PresenceStatusDND, state.Status)
	require.NotNil(t, state.ExpiresAt)

	require.NoError(t, repo.SetEffectiveStatus(ctx, "u1", models.PresenceStatusDND, 60*time.Second))
	require.NoError(t, repo.SetEffectiveStatus(ctx, "u2", models.PresenceStatusIdle, 60*time.Second))
	require.NoError(t, repo.SetEffectiveStatus(ctx, "u3", models.PresenceStatusOffline, 60*time.Second))

	batch, err := repo.BatchGetEffectivePresence(ctx, []string{"u1", "u2", "u3"})
	require.NoError(t, err)
	require.Equal(t, models.UserPresenceSnapshot{IsOnline: true, PresenceStatus: models.PresenceStatusDND}, batch["u1"])
	require.Equal(t, models.UserPresenceSnapshot{IsOnline: true, PresenceStatus: models.PresenceStatusIdle}, batch["u2"])
	require.Equal(t, models.UserPresenceSnapshot{IsOnline: false, PresenceStatus: models.PresenceStatusOffline}, batch["u3"])
}
