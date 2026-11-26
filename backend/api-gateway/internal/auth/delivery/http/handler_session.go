package http

import (
	"net/http"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/api-gateway/pkg/cookie"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

func (h handler) Refresh(w http.ResponseWriter, r *http.Request) {
	// Bind
	token := cookie.RefreshToken(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Service
	out, err := h.ms.User.GRPC.Auth.Refresh(r.Context(), &userv1.RefreshRequest{
		RefreshToken: token,
		UserAgent:    middleware.UserAgent(r),
		ClientIp:     middleware.ClientIP(r),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	authHttp.WriteSession(w, out.GetSession(), h.secureCookie)
}

func (h handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Service
	token := cookie.RefreshToken(r)
	if token != "" {
		if _, err := h.ms.User.GRPC.Auth.Logout(r.Context(), &userv1.LogoutRequest{RefreshToken: token}); err != nil {
			cookie.ClearRefresh(w, h.secureCookie)
			authHttp.WriteOutbound(w, r, h.log, err)
			return
		}
	}

	// Response
	cookie.ClearRefresh(w, h.secureCookie)
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) Me(w http.ResponseWriter, r *http.Request) {
	// Bind
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	// Service
	out, err := h.ms.User.GRPC.Auth.Me(r.Context(), &userv1.MeRequest{AccessToken: token})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.UserFromPB(out.GetUser())))
}
