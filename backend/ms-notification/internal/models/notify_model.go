package models

// SendInput is a notification dispatch request (service layer).
type SendInput struct {
	Channel    string
	TemplateID string
	Locale     string
	Vars       map[string]string
	EmailTo    string
}
