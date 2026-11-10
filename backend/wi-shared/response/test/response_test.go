package response_test

import (
	"testing"

	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/response"
)

func TestEnvelopes(t *testing.T) {
	t.Parallel()

	errEnv := response.NewErrorEnvelope(string(exception.MsgUserNotFound), map[string]any{"id": "1"})
	if errEnv.Error.Code != string(exception.MsgUserNotFound) {
		t.Fatalf("%+v", errEnv)
	}

	dataEnv := response.NewDataEnvelope(map[string]string{"ok": "true"})
	if dataEnv.Data.(map[string]string)["ok"] != "true" {
		t.Fatalf("%+v", dataEnv)
	}
}
