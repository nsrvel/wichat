package test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	presencerepo "github.com/wichat/wichat/backend/ms-user/internal/presence/repository"
	presencesvc "github.com/wichat/wichat/backend/ms-user/internal/presence/service"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	testSupport "github.com/wichat/wichat/backend/ms-user/pkg/test-support"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

func TestPresenceServiceLifecycle(t *testing.T) {
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)

	user := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email:       "presence-user@example.com",
		Password:    "correct-horse-battery",
		DisplayName: "Presence User",
		BirthDate:   "2000-01-01",
	})
	group, err := groupSvc.CreateGroup(ctx, user.AccessToken, "Presence Group")
	require.NoError(t, err)

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	_, groups, sessions := testSupport.TestRepos(st)
	repo := presencerepo.NewRedisRepository(rdb)
	svc := presencesvc.NewService(
		repo,
		groups,
		sessions,
		[]byte(testSupport.TestJWTSecret),
		logger.New(logger.Options{Level: "error"}),
	)

	// 1. Heartbeat active -> online, includes group ID on transition
	tr, err := svc.HeartbeatPresence(ctx, user.AccessToken, "tab-1", models.PresenceActivityActive)
	require.NoError(t, err)
	require.Equal(t, models.PresenceStatusOffline, tr.PreviousStatus)
	require.Equal(t, models.PresenceStatusOnline, tr.EffectiveStatus)
	require.True(t, tr.IsOnline)
	require.Contains(t, tr.GroupIDs, group.ID)

	// 2. Second tab idle while first tab active -> still online
	tr, err = svc.HeartbeatPresence(ctx, user.AccessToken, "tab-2", models.PresenceActivityIdle)
	require.NoError(t, err)
	require.Equal(t, models.PresenceStatusOnline, tr.EffectiveStatus)
	require.Empty(t, tr.GroupIDs)

	// 3. First tab goes idle -> all tabs idle -> effective idle
	tr, err = svc.HeartbeatPresence(ctx, user.AccessToken, "tab-1", models.PresenceActivityIdle)
	require.NoError(t, err)
	require.Equal(t, models.PresenceStatusOnline, tr.PreviousStatus)
	require.Equal(t, models.PresenceStatusIdle, tr.EffectiveStatus)
	require.Contains(t, tr.GroupIDs, group.ID)

	// 4. Manual DND overrides idle
	manualRes, err := svc.SetManualPresence(ctx, user.AccessToken, models.PresenceStatusDND, 3600)
	require.NoError(t, err)
	require.Equal(t, models.PresenceStatusDND, manualRes.ManualStatus)
	require.Equal(t, models.PresenceStatusDND, manualRes.EffectiveStatus)
	require.True(t, manualRes.IsOnline)
	require.NotNil(t, manualRes.ExpiresAt)
	require.Contains(t, manualRes.GroupIDs, group.ID)

	// 5. Manual invisible makes effective status offline while connected
	manualRes, err = svc.SetManualPresence(ctx, user.AccessToken, models.PresenceStatusInvisible, 0)
	require.NoError(t, err)
	require.Equal(t, models.PresenceStatusInvisible, manualRes.ManualStatus)
	require.Equal(t, models.PresenceStatusOffline, manualRes.EffectiveStatus)
	require.False(t, manualRes.IsOnline)

	myView, err := svc.GetMyPresence(ctx, user.AccessToken)
	require.NoError(t, err)
	require.Equal(t, models.PresenceStatusInvisible, myView.ManualStatus)
	require.Equal(t, models.PresenceStatusOffline, myView.EffectiveStatus)

	// 6. Reset manual back to online -> returns to idle (since both tabs idle), then active
	_, err = svc.SetManualPresence(ctx, user.AccessToken, models.PresenceStatusOnline, 0)
	require.NoError(t, err)
	tr, err = svc.HeartbeatPresence(ctx, user.AccessToken, "tab-1", models.PresenceActivityActive)
	require.NoError(t, err)
	require.Equal(t, models.PresenceStatusOnline, tr.EffectiveStatus)

	batch, err := svc.BatchGetEffectivePresence(ctx, []string{user.User.ID, "other-user"})
	require.NoError(t, err)
	require.True(t, batch[user.User.ID].IsOnline)
	require.Equal(t, models.PresenceStatusOnline, batch[user.User.ID].PresenceStatus)
	require.False(t, batch["other-user"].IsOnline)
	require.Equal(t, models.PresenceStatusOffline, batch["other-user"].PresenceStatus)

	// 7. Disconnect tab-1 (tab-2 still idle) -> idle; disconnect tab-2 -> offline
	tr, err = svc.DisconnectPresence(ctx, user.User.ID, "tab-1")
	require.NoError(t, err)
	require.Equal(t, models.PresenceStatusIdle, tr.EffectiveStatus)

	tr, err = svc.DisconnectPresence(ctx, user.User.ID, "tab-2")
	require.NoError(t, err)
	require.Equal(t, models.PresenceStatusOffline, tr.EffectiveStatus)
	require.False(t, tr.IsOnline)
	require.Contains(t, tr.GroupIDs, group.ID)

	// 8. Validation errors
	_, err = svc.HeartbeatPresence(ctx, user.AccessToken, "", models.PresenceActivityActive)
	require.Error(t, err)
	_, err = svc.SetManualPresence(ctx, user.AccessToken, "bogus", 0)
	require.Error(t, err)
	_, err = svc.SetManualPresence(ctx, user.AccessToken, models.PresenceStatusDND, -1)
	require.Error(t, err)
}
