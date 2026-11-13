package username_test

import (
	"context"
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	usermock "github.com/wichat/wichat/backend/ms-user/internal/user/repository/mock"
	"github.com/wichat/wichat/backend/ms-user/pkg/username"
)

func TestBaseFromEmail(t *testing.T) {
	got, err := username.BaseFromEmail("ada@example.com")
	if err != nil || got != "ada" {
		t.Fatalf("ada: %q %v", got, err)
	}

	got, err = username.BaseFromEmail("John.Doe+tag@corp.com")
	if err != nil || got != "john.doe" {
		t.Fatalf("john.doe: %q %v", got, err)
	}

	got, err = username.BaseFromEmail("a@x.com")
	if err != nil || got != "user" {
		t.Fatalf("short local: %q %v", got, err)
	}
}

func TestNormalizeRejectsInvalid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		raw    string
		reason string
	}{
		{"ab", username.ReasonTooShort},
		{strings.Repeat("a", 31), username.ReasonTooLong},
		{"ada!", username.ReasonInvalidFormat},
		{".ada", username.ReasonInvalidFormat},
		{"ada.", username.ReasonInvalidFormat},
		{"_ada", username.ReasonInvalidFormat},
		{"ada_", username.ReasonInvalidFormat},
		{"ada..lovelace", username.ReasonInvalidFormat},
		{"ada__lovelace", username.ReasonInvalidFormat},
		{"ada._lovelace", username.ReasonInvalidFormat},
		{"ada_.lovelace", username.ReasonInvalidFormat},
		{"admin", username.ReasonReserved},
		{"WiChat", username.ReasonReserved},
		{"üser", username.ReasonInvalidFormat},
	}

	for _, tc := range cases {
		_, reason, err := username.Validate(tc.raw)
		if err == nil {
			t.Fatalf("expected error for %q", tc.raw)
		}
		if reason != tc.reason {
			t.Fatalf("%q: got reason %q want %q", tc.raw, reason, tc.reason)
		}
	}

	got, reason, err := username.Validate("  Ada.Lovelace_1  ")
	if err != nil || reason != "" || got != "ada.lovelace_1" {
		t.Fatalf("valid: got %q reason %q err %v", got, reason, err)
	}
}

func TestAllocateUnique(t *testing.T) {
	t.Parallel()

	st := usermock.NewState()
	lookup := usermock.NewPgRepositoryMock(st)
	got, err := username.Allocate(context.Background(), lookup, "fresh@example.com")
	if err != nil || got == "" {
		t.Fatalf("%q %v", got, err)
	}
	st.Users = []models.User{{ID: "u1", Username: got, Email: "fresh@example.com"}}

	got2, err := username.Allocate(context.Background(), lookup, "fresh@example.com")
	if err != nil || got2 == "" || got2 == got {
		t.Fatalf("second %q first %q %v", got2, got, err)
	}
}

func TestAllocateWithManyCollisions(t *testing.T) {
	st := usermock.NewState()
	lookup := usermock.NewPgRepositoryMock(st)
	email := "ada@example.com"

	for i := 0; i < 12; i++ {
		name, err := username.Allocate(context.Background(), lookup, email)
		if err != nil {
			t.Fatal(err)
		}
		st.Users = append(st.Users, models.User{ID: string(rune('a' + i)), Username: name, Email: "x@example.com"})
	}

	last, err := username.Allocate(context.Background(), lookup, email)
	if err != nil || last == "" || len([]rune(last)) > username.MaxLen {
		t.Fatalf("%q %v", last, err)
	}
}

func TestBaseFromEmailSanitizesSpecialChars(t *testing.T) {
	got, err := username.BaseFromEmail("!!!weird!!!@example.com")
	if err != nil || got != "weird" {
		t.Fatalf("%q %v", got, err)
	}
}

func TestAllocateAddsNumericSuffix(t *testing.T) {
	st := usermock.NewState()
	lookup := usermock.NewPgRepositoryMock(st)
	email := "ada@example.com"
	first, err := username.Allocate(context.Background(), lookup, email)
	if err != nil {
		t.Fatal(err)
	}
	st.Users = append(st.Users, models.User{ID: "u1", Username: first, Email: email})

	second, err := username.Allocate(context.Background(), lookup, email)
	if err != nil || second == first {
		t.Fatalf("%q %q %v", first, second, err)
	}
	if !strings.HasSuffix(second, "2") {
		t.Fatalf("suffix %q", second)
	}
}
