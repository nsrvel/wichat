package service_test

import (
	"context"
	"testing"

	"github.com/wichat/wichat/backend/ms-notification/internal/models"
	"github.com/wichat/wichat/backend/ms-notification/internal/notify/service"
	"github.com/wichat/wichat/backend/ms-notification/pkg/constants"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
)

type fakeMail struct {
	sent int
}

func (f *fakeMail) SendHTML(context.Context, string, string, string) error {
	f.sent++
	return nil
}

func TestSendRejectsUnknownTemplate(t *testing.T) {
	svc := service.NewService(&fakeMail{}, logger.New(logger.Options{Service: "test", Env: "test", Level: "error"}))
	err := svc.Send(context.Background(), models.SendInput{
		Channel:    constants.NotificationChannelEmail,
		TemplateID: "unknown",
		EmailTo:    "a@b.com",
	})
	if codeOf(t, err) != string(exception.MsgInvalidInput) {
		t.Fatalf("err: %v", err)
	}
}

func TestSendPasswordReset(t *testing.T) {
	mail := &fakeMail{}
	svc := service.NewService(mail, logger.New(logger.Options{Service: "test", Env: "test", Level: "error"}))
	err := svc.Send(context.Background(), models.SendInput{
		Channel:    constants.NotificationChannelEmail,
		TemplateID: "auth.password_reset",
		Locale:     "en",
		EmailTo:    "a@b.com",
		Vars: map[string]string{
			"code":            "ABCD12",
			"display_name":    "Ada",
			"expires_minutes": "10",
		},
	})
	if err != nil || mail.sent != 1 {
		t.Fatalf("err=%v sent=%d", err, mail.sent)
	}
}

func codeOf(t *testing.T, err error) string {
	t.Helper()
	ex, ok := exception.As(err)
	if !ok {
		t.Fatalf("not exception: %v", err)
	}
	return string(ex.Code)
}
