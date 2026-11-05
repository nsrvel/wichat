package token

import (
	"regexp"
	"testing"
)

func TestNewPublicIDFormat(t *testing.T) {
	id, err := NewPublicID()
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^[0-9a-f]{24}$`).MatchString(id) {
		t.Fatalf("id %q", id)
	}
}
