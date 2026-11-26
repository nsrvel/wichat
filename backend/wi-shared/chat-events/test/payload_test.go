package chatEvents_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	chatEvents "github.com/wichat/wichat/backend/wi-shared/chat-events"
)

func TestNormalizePayload_object(t *testing.T) {
	raw := []byte(`{"type":"message.created","group_id":"g"}`)
	got, err := chatEvents.NormalizePayload(raw)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(raw) {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizePayload_jsonStringBody(t *testing.T) {
	inner := []byte(`{"type":"message.updated","group_id":"g"}`)
	wrapped, err := json.Marshal(string(inner))
	if err != nil {
		t.Fatal(err)
	}
	got, err := chatEvents.NormalizePayload(wrapped)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(inner) {
		t.Fatalf("got %q want %q", got, inner)
	}
}

func TestNormalizePayload_legacyBase64String(t *testing.T) {
	inner := []byte(`{"type":"message.deleted","group_id":"g","channel_id":"c"}`)
	b64 := base64.StdEncoding.EncodeToString(inner)
	wrapped, err := json.Marshal(b64)
	if err != nil {
		t.Fatal(err)
	}
	got, err := chatEvents.NormalizePayload(wrapped)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(inner) {
		t.Fatalf("got %q want %q", got, inner)
	}
}

func TestNormalizePayload_errors(t *testing.T) {
	_, err := chatEvents.NormalizePayload(nil)
	if err == nil {
		t.Fatal("expected error for empty")
	}
	_, err = chatEvents.NormalizePayload([]byte(`"not-json-object"`))
	if err == nil {
		t.Fatal("expected error for non-object string wrapper")
	}
	_, err = chatEvents.NormalizePayload([]byte("{not-valid"))
	if err == nil {
		t.Fatal("expected error for invalid object json")
	}
}
