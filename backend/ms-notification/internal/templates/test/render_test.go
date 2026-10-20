package templates_test

import (
	"strings"
	"testing"

	"github.com/wichat/wichat/backend/ms-notification/internal/templates"
)

func TestRenderPasswordResetEn(t *testing.T) {
	subject, html, err := templates.RenderPasswordReset("en", "Ada", "ABC123", "10")
	if err != nil {
		t.Fatal(err)
	}
	if subject != "Your WiChat password reset code" {
		t.Fatalf("subject: %q", subject)
	}
	if !strings.Contains(html, "ABC-123") {
		t.Fatal("html missing formatted code")
	}
}

func TestRenderPasswordResetId(t *testing.T) {
	subject, html, err := templates.RenderPasswordReset("id", "Budi", "XYZ789", "10")
	if err != nil {
		t.Fatal(err)
	}
	if subject != "Kode reset kata sandi WiChat Anda" {
		t.Fatalf("subject: %q", subject)
	}
	if !strings.Contains(html, "XYZ-789") {
		t.Fatal("missing formatted code")
	}
}
