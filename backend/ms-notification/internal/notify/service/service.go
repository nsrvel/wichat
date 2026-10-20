package service

import (
	"context"
	"strings"

	"github.com/wichat/wichat/backend/ms-notification/internal/models"
	"github.com/wichat/wichat/backend/ms-notification/internal/templates"
	"github.com/wichat/wichat/backend/ms-notification/pkg/constants"
	"github.com/wichat/wichat/backend/ms-notification/pkg/mail"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

// Service dispatches outbound notifications.
type Service interface {
	Send(ctx context.Context, in models.SendInput) error
}

type service struct {
	mail mail.Sender
	log  logger.Logger
}

// NewService builds the notify use cases.
func NewService(mailSender mail.Sender, log logger.Logger) Service {
	return service{mail: mailSender, log: log}
}

func (s service) Send(ctx context.Context, in models.SendInput) error {
	// Validation
	if strings.TrimSpace(in.Channel) != constants.NotificationChannelEmail {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "channel"})
	}
	to := strings.TrimSpace(in.EmailTo)
	if to == "" {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "email.to"})
	}

	switch strings.TrimSpace(in.TemplateID) {
	case templates.TemplatePasswordReset:
		return s.sendPasswordReset(ctx, to, in.Locale, in.Vars)
	case templates.TemplateEmailCode:
		return s.sendEmailCode(ctx, to, in.Locale, in.Vars)
	default:
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "template_id"})
	}
}

func (s service) sendPasswordReset(ctx context.Context, to, locale string, vars map[string]string) error {
	code := vars["code"]
	displayName := vars["display_name"]
	expires := vars["expires_minutes"]
	if strings.TrimSpace(code) == "" {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "vars.code"})
	}

	subject, html, err := templates.RenderPasswordReset(locale, displayName, code, expires)
	if err != nil {
		return exception.Internal(err)
	}

	if err := s.mail.SendHTML(ctx, to, subject, html); err != nil {
		s.log.Error("smtp send failed", "err", err, "template", templates.TemplatePasswordReset)
		return exception.Unavailable(exception.MsgUnavailable, nil)
	}

	return nil
}

func (s service) sendEmailCode(ctx context.Context, to, locale string, vars map[string]string) error {
	code := vars["code"]
	displayName := vars["display_name"]
	expires := vars["expires_minutes"]
	purpose := vars["purpose"]
	if strings.TrimSpace(code) == "" {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "vars.code"})
	}

	subject, html, err := templates.RenderEmailCode(locale, displayName, code, expires, purpose)
	if err != nil {
		return exception.Internal(err)
	}

	if err := s.mail.SendHTML(ctx, to, subject, html); err != nil {
		s.log.Error("smtp send failed", "err", err, "template", templates.TemplateEmailCode)
		return exception.Unavailable(exception.MsgUnavailable, nil)
	}

	return nil
}
