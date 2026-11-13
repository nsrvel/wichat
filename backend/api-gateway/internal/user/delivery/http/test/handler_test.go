package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	userdelivery "github.com/wichat/wichat/backend/api-gateway/internal/user/delivery/http"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	msUser "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeUser struct {
	setupStatus func(context.Context, *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error)
}

func (f fakeUser) SetupStatus(ctx context.Context, in *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error) {
	if f.setupStatus != nil {
		return f.setupStatus(ctx, in)
	}

	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (fakeUser) CompleteInstanceSetup(context.Context, *userv1.CompleteInstanceSetupRequest) (*userv1.CompleteInstanceSetupResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func (fakeUser) SearchUsers(context.Context, *userv1.SearchUsersRequest) (*userv1.SearchUsersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "error.internal")
}

func TestSetupStatusReturnsInstancePolicy(t *testing.T) {
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		User: fakeUser{
			setupStatus: func(context.Context, *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error) {
				return &userv1.SetupStatusResponse{
					RegistrationMode: "invite_only",
					UsernameEnabled:  false,
				}, nil
			},
		},
	}}}
	handler := userdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/setup-status", nil)
	rec := httptest.NewRecorder()
	handler.SetupStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data struct {
			RegistrationMode string `json:"registration_mode"`
			UsernameEnabled  bool   `json:"username_enabled"`
			NeedsAccount     *bool  `json:"needs_account"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.RegistrationMode != "invite_only" || body.Data.UsernameEnabled {
		t.Fatalf("policy: %+v", body.Data)
	}
	if body.Data.NeedsAccount != nil {
		t.Fatalf("deprecated field leaked: %+v", body.Data)
	}
}
