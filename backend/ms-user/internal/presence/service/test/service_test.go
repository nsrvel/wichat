package test

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

	mr.Del("presence:effective:" + user.User.ID)
	batch, err = svc.BatchGetEffectivePresence(ctx, []string{user.User.ID})
	require.NoError(t, err)
	require.False(t, batch[user.User.ID].IsOnline)
	require.Equal(t, models.PresenceStatusOffline, batch[user.User.ID].PresenceStatus)

	mr.Del("presence:effective:" + user.User.ID)
	myView, err = svc.GetMyPresence(ctx, user.AccessToken)
	require.NoError(t, err)
	require.True(t, myView.IsOnline)
	require.Equal(t, models.PresenceStatusOnline, myView.EffectiveStatus)

	// 7. Disconnect tab-1 (tab-2 still idle) -> idle; disconnect tab-2 -> offline
	tr, err = svc.DisconnectPresence(ctx, user.User.ID, "tab-1")
	require.NoError(t, err)
	require.Equal(t, models.PresenceStatusIdle, tr.EffectiveStatus)

	tr, err = svc.DisconnectPresence(ctx, user.User.ID, "tab-2")
	require.NoError(t, err)
	require.Equal(t, models.PresenceStatusOffline, tr.EffectiveStatus)
	require.False(t, tr.IsOnline)
	require.Contains(t, tr.GroupIDs, group.ID)

	// 8. Resolve-only read when effective key expired but conns remain
	tr, err = svc.HeartbeatPresence(ctx, user.AccessToken, "tab-resolve", models.PresenceActivityActive)
	require.NoError(t, err)
	require.Equal(t, models.PresenceStatusOnline, tr.EffectiveStatus)
	mr.Del("presence:effective:" + user.User.ID)
	myView, err = svc.GetMyPresence(ctx, user.AccessToken)
	require.NoError(t, err)
	require.True(t, myView.IsOnline)
	batch, err = svc.BatchGetEffectivePresence(ctx, []string{user.User.ID})
	require.NoError(t, err)
	require.True(t, batch[user.User.ID].IsOnline)

	// 9. Validation errors
	_, err = svc.HeartbeatPresence(ctx, user.AccessToken, "", models.PresenceActivityActive)
	require.Error(t, err)
	_, err = svc.SetManualPresence(ctx, user.AccessToken, "bogus", 0)
	require.Error(t, err)
	_, err = svc.SetManualPresence(ctx, user.AccessToken, models.PresenceStatusDND, -1)
	require.Error(t, err)
}

func newPresenceSvc(t *testing.T) (context.Context, presencesvc.Service, *models.SessionResult, func()) {
	t.Helper()
	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	user := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
		Email:       fmt.Sprintf("presence-%s@example.com", t.Name()),
		Password:    "correct-horse-battery",
		DisplayName: "Presence Edge",
		BirthDate:   "2000-01-01",
	})
	_, err := groupSvc.CreateGroup(ctx, user.AccessToken, "Edge Group")
	require.NoError(t, err)

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	_, groups, sessions := testSupport.TestRepos(st)
	repo := presencerepo.NewRedisRepository(rdb)
	svc := presencesvc.NewService(
		repo,
		groups,
		sessions,
		[]byte(testSupport.TestJWTSecret),
		logger.New(logger.Options{Level: "error"}),
	)
	cleanup := func() {
		_ = rdb.Close()
	}
	return ctx, svc, user, cleanup
}

func TestMultiDeviceOneDisconnectStaysOnlineForOthers(t *testing.T) {
	ctx, svc, user, cleanup := newPresenceSvc(t)
	defer cleanup()

	_, err := svc.HeartbeatPresence(ctx, user.AccessToken, "phone", models.PresenceActivityActive)
	require.NoError(t, err)
	_, err = svc.HeartbeatPresence(ctx, user.AccessToken, "desktop", models.PresenceActivityActive)
	require.NoError(t, err)

	tr, err := svc.DisconnectPresence(ctx, user.User.ID, "phone")
	require.NoError(t, err)
	require.True(t, tr.IsOnline)
	require.Equal(t, models.PresenceStatusOnline, tr.EffectiveStatus)

	batch, err := svc.BatchGetEffectivePresence(ctx, []string{user.User.ID})
	require.NoError(t, err)
	require.True(t, batch[user.User.ID].IsOnline)
	require.Equal(t, models.PresenceStatusOnline, batch[user.User.ID].PresenceStatus)
}

func TestMultiDeviceActivityMergeIdleAndActive(t *testing.T) {
	ctx, svc, user, cleanup := newPresenceSvc(t)
	defer cleanup()

	_, err := svc.HeartbeatPresence(ctx, user.AccessToken, "a", models.PresenceActivityActive)
	require.NoError(t, err)
	_, err = svc.HeartbeatPresence(ctx, user.AccessToken, "b", models.PresenceActivityIdle)
	require.NoError(t, err)

	tr, err := svc.DisconnectPresence(ctx, user.User.ID, "a")
	require.NoError(t, err)
	require.True(t, tr.IsOnline)
	require.Equal(t, models.PresenceStatusIdle, tr.EffectiveStatus)

	_, err = svc.HeartbeatPresence(ctx, user.AccessToken, "c", models.PresenceActivityActive)
	require.NoError(t, err)
	batch, err := svc.BatchGetEffectivePresence(ctx, []string{user.User.ID})
	require.NoError(t, err)
	require.Equal(t, models.PresenceStatusOnline, batch[user.User.ID].PresenceStatus)
}

func TestPresenceDegradesWhenRedisUnavailable(t *testing.T) {
	ctx, svc, user, cleanup := newPresenceSvc(t)
	defer cleanup()

	_, err := svc.HeartbeatPresence(ctx, user.AccessToken, "tab", models.PresenceActivityActive)
	require.NoError(t, err)

	cleanup()

	_, err = svc.HeartbeatPresence(ctx, user.AccessToken, "tab", models.PresenceActivityActive)
	require.Error(t, err)

	view, err := svc.GetMyPresence(ctx, user.AccessToken)
	require.NoError(t, err)
	require.Equal(t, models.PresenceStatusOnline, view.ManualStatus)
	require.Equal(t, models.PresenceStatusOffline, view.EffectiveStatus)
	require.False(t, view.IsOnline)

	batch, err := svc.BatchGetEffectivePresence(ctx, []string{user.User.ID, "other"})
	require.NoError(t, err)
	require.False(t, batch[user.User.ID].IsOnline)
	require.Equal(t, models.PresenceStatusOffline, batch[user.User.ID].PresenceStatus)
	require.False(t, batch["other"].IsOnline)
}

func TestPresenceHeartbeatLoadSample(t *testing.T) {
	if testing.Short() {
		t.Skip("load sample")
	}
	ctx, svc, user, cleanup := newPresenceSvc(t)
	defer cleanup()

	const (
		heartbeats = 3_000
		conns      = 32
	)
	start := time.Now()
	for i := 0; i < heartbeats; i++ {
		conn := fmt.Sprintf("load-conn-%d", i%conns)
		activity := models.PresenceActivityActive
		if i%5 == 0 {
			activity = models.PresenceActivityIdle
		}
		if _, err := svc.HeartbeatPresence(ctx, user.AccessToken, conn, activity); err != nil {
			t.Fatal(err)
		}
	}
	elapsed := time.Since(start)
	t.Logf("ms-user HeartbeatPresence: %d heartbeats in %v (%.0f ops/s, %d logical conns)",
		heartbeats, elapsed, float64(heartbeats)/elapsed.Seconds(), conns)
}

func presencePercentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(p*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func TestPresenceRealisticLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("load sample")
	}

	ctx := context.Background()
	st := usermock.NewState()
	authSvc, notify, _, _, groupSvc, _ := testSupport.NewAuthService(t, st, nil)
	_, groups, sessions := testSupport.TestRepos(st)

	const (
		users        = 12
		connsPerUser = 4
		beatsPerConn = 40
		jitterMaxMs  = 3
	)

	type account struct {
		token string
		id    string
	}
	accounts := make([]account, 0, users)
	for i := 0; i < users; i++ {
		u := testSupport.RegisterSession(t, authSvc, notify, models.RegisterInput{
			Email:       fmt.Sprintf("load-user-%d@example.com", i),
			Password:    "correct-horse-battery",
			DisplayName: fmt.Sprintf("Load %d", i),
			BirthDate:   "2000-01-01",
		})
		if _, err := groupSvc.CreateGroup(ctx, u.AccessToken, fmt.Sprintf("G%d", i)); err != nil {
			t.Fatalf("CreateGroup: %v", err)
		}
		accounts = append(accounts, account{token: u.AccessToken, id: u.User.ID})
	}

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	repo := presencerepo.NewRedisRepository(rdb)
	svc := presencesvc.NewService(
		repo,
		groups,
		sessions,
		[]byte(testSupport.TestJWTSecret),
		logger.New(logger.Options{Level: "error"}),
	)

	var (
		latMu     sync.Mutex
		latencies []time.Duration
		errorsN   atomic.Int32
	)
	record := func(d time.Duration, err error) {
		if err != nil {
			errorsN.Add(1)
			return
		}
		latMu.Lock()
		latencies = append(latencies, d)
		latMu.Unlock()
	}

	start := time.Now()
	var wg sync.WaitGroup
	for ui, acct := range accounts {
		for ci := 0; ci < connsPerUser; ci++ {
			wg.Add(1)
			connID := fmt.Sprintf("u%d-tab%d", ui, ci)
			go func(token, conn string, userIdx, connIdx int) {
				defer wg.Done()
				for beat := 0; beat < beatsPerConn; beat++ {
					if beat == beatsPerConn/2 {
						_, _ = svc.DisconnectPresence(ctx, accounts[userIdx].id, conn)
						time.Sleep(time.Duration((beat+userIdx)%jitterMaxMs) * time.Millisecond)
						conn = fmt.Sprintf("%s-r%d", conn, beat)
					}
					time.Sleep(time.Duration((beat+userIdx+connIdx)%jitterMaxMs) * time.Millisecond)
					activity := models.PresenceActivityActive
					if beat%7 == 0 {
						activity = models.PresenceActivityIdle
					}
					t0 := time.Now()
					_, err := svc.HeartbeatPresence(ctx, token, conn, activity)
					record(time.Since(t0), err)
				}
			}(acct.token, connID, ui, ci)
		}
	}
	wg.Wait()
	elapsed := time.Since(start)

	latMu.Lock()
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	sorted := append([]time.Duration(nil), latencies...)
	latMu.Unlock()

	totalOps := users * connsPerUser * beatsPerConn
	errRate := float64(errorsN.Load()) / float64(totalOps) * 100
	t.Logf("ms-user realistic load: users=%d conns/user=%d beats/conn=%d total_ops=%d duration=%v throughput=%.0f ops/s errors=%d (%.2f%%)",
		users, connsPerUser, beatsPerConn, totalOps, elapsed, float64(totalOps)/elapsed.Seconds(), errorsN.Load(), errRate)
	maxLat := time.Duration(0)
	if len(sorted) > 0 {
		maxLat = sorted[len(sorted)-1]
	}
	t.Logf("latency ms-user: p50=%v p95=%v p99=%v max=%v",
		presencePercentile(sorted, 0.50),
		presencePercentile(sorted, 0.95),
		presencePercentile(sorted, 0.99),
		maxLat,
	)
	require.Equal(t, int32(0), errorsN.Load())
}
