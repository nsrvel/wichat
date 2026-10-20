package templates

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"regexp"
	"strings"
)

//go:embed en/*.html id/*.html
var files embed.FS

var titleRe = regexp.MustCompile(`(?is)<title[^>]*>([^<]*)</title>`)

const (
	TemplatePasswordReset = "auth.password_reset"
	TemplateEmailCode     = "auth.email_code"
)

// PasswordResetView is template data for auth.password_reset.
type PasswordResetView struct {
	DisplayName    string
	CodeFormatted  string
	ExpiresMinutes string
}

// EmailCodeView is template data for auth.email_code.
type EmailCodeView struct {
	DisplayName   string
	CodeFormatted string
	Headline      string
	Lead          string
	SubjectTitle  string
}

// RenderEmailCode returns subject and HTML for sign-up or sign-in OTP mail.
func RenderEmailCode(locale, displayName, code, purpose string) (subject, html string, err error) {
	loc := normalizeLocale(locale)
	htmlBytes, err := files.ReadFile(fmt.Sprintf("%s/auth_email_code.html", loc))
	if err != nil {
		return "", "", err
	}

	headline, lead, subjectTitle := emailCodeCopy(loc, purpose)
	view := EmailCodeView{
		DisplayName:   strings.TrimSpace(displayName),
		CodeFormatted: formatCode(code),
		Headline:      headline,
		Lead:          lead,
		SubjectTitle:  subjectTitle,
	}
	if view.DisplayName == "" {
		view.DisplayName = "there"
	}

	html, err = executeTemplate(string(htmlBytes), view)
	if err != nil {
		return "", "", err
	}

	subject, err = subjectFromHTML(html)
	if err != nil {
		return "", "", err
	}

	return subject, html, nil
}

func emailCodeCopy(locale, purpose string) (headline, lead, subjectTitle string) {
	if normalizeLocale(locale) == "id" {
		switch purpose {
		case "sign_up":
			return "Verifikasi email Anda", "gunakan kode ini untuk menyelesaikan pendaftaran.", "Kode verifikasi WiChat"
		default:
			return "Verifikasi masuk", "gunakan kode ini untuk menyelesaikan masuk.", "Kode masuk WiChat"
		}
	}
	switch purpose {
	case "sign_up":
		return "Verify your email", "use this code to finish creating your account.", "Your WiChat verification code"
	default:
		return "Verify your sign-in", "use this code to finish signing in.", "Your WiChat sign-in code"
	}
}

func RenderPasswordReset(locale, displayName, code, expiresMinutes string) (subject, html string, err error) {
	loc := normalizeLocale(locale)
	htmlBytes, err := files.ReadFile(fmt.Sprintf("%s/auth_password_reset.html", loc))
	if err != nil {
		return "", "", err
	}

	view := PasswordResetView{
		DisplayName:    strings.TrimSpace(displayName),
		CodeFormatted:  formatCode(code),
		ExpiresMinutes: strings.TrimSpace(expiresMinutes),
	}
	if view.DisplayName == "" {
		view.DisplayName = "there"
	}
	if view.ExpiresMinutes == "" {
		view.ExpiresMinutes = "10"
	}

	html, err = executeTemplate(string(htmlBytes), view)
	if err != nil {
		return "", "", err
	}

	subject, err = subjectFromHTML(html)
	if err != nil {
		return "", "", err
	}

	return subject, html, nil
}

func subjectFromHTML(html string) (string, error) {
	match := titleRe.FindStringSubmatch(html)
	if len(match) < 2 {
		return "", fmt.Errorf("templates: missing <title>")
	}

	return strings.TrimSpace(match[1]), nil
}

func normalizeLocale(locale string) string {
	switch strings.ToLower(strings.TrimSpace(locale)) {
	case "id":
		return "id"
	default:
		return "en"
	}
}

func formatCode(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) == 6 {
		return code[:3] + "-" + code[3:]
	}

	return code
}

func executeTemplate(body string, data any) (string, error) {
	tmpl, err := template.New("email").Parse(body)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
