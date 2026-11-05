package mail

import (
	"context"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// LogSender logs HTML email in dev instead of delivering (ponytail: local auth OTP without SMTP).
type LogSender struct {
	log logger.Logger
}

// NewLogSender writes outbound mail to service logs.
func NewLogSender(log logger.Logger) Sender {
	return LogSender{log: log}
}

func (s LogSender) SendHTML(ctx context.Context, to, subject, htmlBody string) error {
	s.log.Info("dev mail (not sent)",
		"to", to,
		"subject", subject,
		"html", htmlBody,
	)

	return nil
}
