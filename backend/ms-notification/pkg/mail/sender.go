package mail

import (
	"context"
	"fmt"
	"strings"

	"github.com/wichat/wichat/backend/ms-notification/internal/config"
	mail "github.com/wneessen/go-mail"
)

// Sender delivers HTML email over SMTP.
type Sender interface {
	SendHTML(ctx context.Context, to, subject, htmlBody string) error
}

type smtpSender struct {
	from   string
	client *mail.Client
}

// NewSender builds an SMTP sender from service config.
func NewSender(cfg config.Config) (Sender, error) {
	from := strings.TrimSpace(cfg.SMTPFrom)
	if from == "" {
		return nil, fmt.Errorf("mail: SMTP_FROM is required")
	}

	opts := []mail.Option{
		mail.WithPort(cfg.SMTPPort),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(cfg.SMTPUser),
		mail.WithPassword(cfg.SMTPPassword),
	}
	if cfg.SMTPSecure {
		opts = append(opts, mail.WithSSL())
	} else {
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	}

	client, err := mail.NewClient(cfg.SMTPHost, opts...)
	if err != nil {
		return nil, err
	}

	return smtpSender{from: from, client: client}, nil
}

func (s smtpSender) SendHTML(ctx context.Context, to, subject, htmlBody string) error {
	msg := mail.NewMsg()
	if err := msg.From(s.from); err != nil {
		return err
	}
	if err := msg.To(strings.TrimSpace(to)); err != nil {
		return err
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextHTML, htmlBody)

	return s.client.DialAndSendWithContext(ctx, msg)
}
