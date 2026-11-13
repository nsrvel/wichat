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
	setupStatus   func(context.Context, *userv1.SetupStatusRequest) (*userv1.SetupStatusResponse, error)
	checkUsername func(context.Context, *userv1.CheckUsernameRequest) (*userv1.CheckUsernameResponse, error)
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

func (f fakeUser) CheckUsername(ctx context.Context, in *userv1.CheckUsernameRequest) (*userv1.CheckUsernameResponse, error) {
	if f.checkUsername != nil {
		return f.checkUsername(ctx, in)
	}

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
			NeedsAccount     *bool  `json:"needs_account"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.RegistrationMode != "invite_only" {
		t.Fatalf("policy: %+v", body.Data)
	}
	if body.Data.NeedsAccount != nil {
		t.Fatalf("deprecated field leaked: %+v", body.Data)
	}
}

func TestCheckUsernameReturnsAvailability(t *testing.T) {
	clients := &microservices.Clients{User: msUser.Service{GRPC: msUser.GRPC{
		User: fakeUser{
			checkUsername: func(_ context.Context, in *userv1.CheckUsernameRequest) (*userv1.CheckUsernameResponse, error) {
				return &userv1.CheckUsernameResponse{
					Available:  true,
					Normalized: "ada.lovelace",
				}, nil
			},
		},
	}}}
	handler := userdelivery.NewHandler(clients, logger.New(logger.Options{Service: "gateway-test", Env: "test", Level: "error"}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/check-username?username=Ada.Lovelace", nil)
	rec := httptest.NewRecorder()
	handler.CheckUsername(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}
