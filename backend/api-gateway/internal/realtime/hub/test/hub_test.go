package hub_test

import (
	"testing"

	"github.com/wichat/wichat/backend/api-gateway/internal/realtime/hub"
)

func TestDeliver_noSubscribers_noPanic(t *testing.T) {
	h := hub.New()
	h.Deliver("aaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbb", []byte(`{"type":"message.created"}`))
}
