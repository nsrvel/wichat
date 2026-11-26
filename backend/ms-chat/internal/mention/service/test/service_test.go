package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/uptrace/bun"
	mentionservice "github.com/wichat/wichat/backend/ms-chat/internal/mention/service"
	"github.com/wichat/wichat/backend/ms-chat/internal/models"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	userauth "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user/grpc/auth"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

type fakeChannel struct{}

func (fakeChannel) ListGroupChannels(_ context.Context, _ *userv1.ListGroupChannelsRequest) (*userv1.ListGroupChannelsResponse, error) {
	return &userv1.ListGroupChannelsResponse{
		Channels: []*userv1.ChannelSummary{
			{Id: "ch-1"},
		},
	}, nil
}

type fakeUserAuth struct{}

func (fakeUserAuth) Setup(context.Context, *userv1.SetupRequest) (*userv1.SetupResponse, error) {
	return nil, nil
}

func (fakeUserAuth) Register(context.Context, *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	return nil, nil
}

func (fakeUserAuth) Login(context.Context, *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	return nil, nil
}

func (fakeUserAuth) VerifyRegister(context.Context, *userv1.VerifyRegisterRequest) (*userv1.VerifyRegisterResponse, error) {
	return nil, nil
}

func (fakeUserAuth) VerifyLogin(context.Context, *userv1.VerifyLoginRequest) (*userv1.VerifyLoginResponse, error) {
	return nil, nil
}

func (fakeUserAuth) ResendRegisterCode(context.Context, *userv1.ResendRegisterCodeRequest) (*userv1.ResendRegisterCodeResponse, error) {
	return nil, nil
}

func (fakeUserAuth) ResendLoginCode(context.Context, *userv1.ResendLoginCodeRequest) (*userv1.ResendLoginCodeResponse, error) {
	return nil, nil
}

func (fakeUserAuth) Refresh(context.Context, *userv1.RefreshRequest) (*userv1.RefreshResponse, error) {
	return nil, nil
}

func (fakeUserAuth) Logout(context.Context, *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	return nil, nil
}

func (fakeUserAuth) Me(context.Context, *userv1.MeRequest) (*userv1.MeResponse, error) {
	return &userv1.MeResponse{User: &userv1.User{Id: "u-pub"}}, nil
}

func (fakeUserAuth) GoogleStart(context.Context, *userv1.GoogleStartRequest) (*userv1.GoogleStartResponse, error) {
	return nil, nil
}

func (fakeUserAuth) GoogleFinish(context.Context, *userv1.GoogleFinishRequest) (*userv1.GoogleFinishResponse, error) {
	return nil, nil
}

func (fakeUserAuth) RequestPasswordReset(context.Context, *userv1.RequestPasswordResetRequest) (*userv1.RequestPasswordResetResponse, error) {
	return nil, nil
}

func (fakeUserAuth) VerifyPasswordReset(context.Context, *userv1.VerifyPasswordResetRequest) (*userv1.VerifyPasswordResetResponse, error) {
	return nil, nil
}

func (fakeUserAuth) CompletePasswordReset(context.Context, *userv1.CompletePasswordResetRequest) (*userv1.CompletePasswordResetResponse, error) {
	return nil, nil
}

func (fakeUserAuth) VerifyLoginMFA(context.Context, *userv1.VerifyLoginMFARequest) (*userv1.VerifyLoginMFAResponse, error) {
	return nil, nil
}

func (fakeUserAuth) BeginLoginWebAuthn(context.Context, *userv1.BeginLoginWebAuthnRequest) (*userv1.BeginLoginWebAuthnResponse, error) {
	return nil, nil
}

var _ userauth.Client = fakeUserAuth{}

type fakeMentionsRepo struct {
	hits []models.GroupMentionHit
}

func (fakeMentionsRepo) InsertMentions(context.Context, bun.Tx, []models.MessageMentionRow) error {
	return nil
}

func (fakeMentionsRepo) ResolveUserPublicIDsForChannel(context.Context, string, []string) ([]string, error) {
	return nil, nil
}

func (fakeMentionsRepo) LookupGroupInternalID(context.Context, string) (string, error) {
	return "g-int", nil
}

func (fakeMentionsRepo) LookupUserInternalID(context.Context, string) (string, error) {
	return "u-int", nil
}

func (f fakeMentionsRepo) ListForMentionedUser(
	context.Context, string, string, int, time.Time, string, bool,
) ([]models.GroupMentionHit, error) {
	return f.hits, nil
}

func (fakeMentionsRepo) MarkRead(context.Context, string, string, []string, time.Time) error {
	return nil
}

func (fakeMentionsRepo) CountUnread(context.Context, string, string) (int, error) {
	return 0, nil
}

func (fakeMentionsRepo) MentionedPublicUserIDsByMessageIDs(context.Context, []string) (map[string][]string, error) {
	return map[string][]string{}, nil
}

func TestListGroupMentionsFiltersInaccessibleChannel(t *testing.T) {
	mentions := fakeMentionsRepo{hits: []models.GroupMentionHit{
		{MessageID: "m1", ChannelID: "ch-1", CreatedAt: time.Now().UTC()},
		{MessageID: "m2", ChannelID: "ch-secret", CreatedAt: time.Now().UTC()},
	}}
	svc := mentionservice.NewService(mentions, fakeChannel{}, fakeUserAuth{}, logger.New(logger.Options{}))

	hits, _, err := svc.ListGroupMentions(context.Background(), &models.ListGroupMentionsInput{
		AccessToken:   "token",
		GroupPublicID: "g1",
		Limit:         8,
	})
	if err != nil {
		t.Fatalf("ListGroupMentions: %v", err)
	}
	if len(hits) != 1 || hits[0].MessageID != "m1" {
		t.Fatalf("expected one accessible mention hit, got %+v", hits)
	}
}
