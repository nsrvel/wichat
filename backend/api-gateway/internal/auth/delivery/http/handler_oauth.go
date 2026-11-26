package http

import (
	"net/http"
	"strings"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/api-gateway/pkg/cookie"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

func (h handler) GoogleStart(w http.ResponseWriter, r *http.Request) {
	if !h.verifyTurnstile(w, r, r.URL.Query().Get("turnstile_token")) {
		return
	}

	if h.inviteOnly && authHttp.OAuthSignupReturnBlocked(r.URL.Query().Get("next")) {
		authHttp.RedirectForbidden(w, r, h.webOrigin)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.Auth.GoogleStart(r.Context(), &userv1.GoogleStartRequest{})
	if err != nil {
		authHttp.RedirectSignIn(w, r, h.webOrigin, authHttp.GRPCMessageCode(err))
		return
	}
	if out.GetAuthorizationUrl() == "" || out.GetState() == "" || out.GetCodeVerifier() == "" {
		authHttp.RedirectSignIn(w, r, h.webOrigin, string(exception.MsgUnavailable))
		return
	}

	// Response
	cookie.SetOAuth(w, out.GetState(), out.GetCodeVerifier(), h.secureCookie)
	cookie.SetOAuthReturn(w, authHttp.SafeReturnPath(r.URL.Query().Get("next")), h.secureCookie)
	http.Redirect(w, r, out.GetAuthorizationUrl(), http.StatusFound)
}

func (h handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Bind
	cookie.ClearOAuth(w, h.secureCookie)
	if r.URL.Query().Get("error") != "" {
		authHttp.RedirectSignIn(w, r, h.webOrigin, string(exception.MsgGoogleDenied))
		return
	}

	state, verifier := cookie.OAuthStateVerifier(r)
	code := r.URL.Query().Get("code")
	if state == "" || verifier == "" || code == "" || state != r.URL.Query().Get("state") {
		authHttp.RedirectSignIn(w, r, h.webOrigin, string(exception.MsgGoogleDenied))
		return
	}

	// Service
	out, err := h.ms.User.GRPC.Auth.GoogleFinish(r.Context(), &userv1.GoogleFinishRequest{
		Code:         code,
		CodeVerifier: verifier,
		UserAgent:    middleware.UserAgent(r),
		ClientIp:     middleware.ClientIP(r),
	})
	if err != nil {
		code := authHttp.GRPCMessageCode(err)
		if code == string(exception.MsgRegistrationClosed) {
			authHttp.RedirectForbidden(w, r, h.webOrigin)
			return
		}
		authHttp.RedirectSignIn(w, r, h.webOrigin, code)
		return
	}
	sess := out.GetSession()
	if sess == nil || sess.GetRefreshToken() == "" {
		authHttp.RedirectSignIn(w, r, h.webOrigin, string(exception.MsgInternal))
		return
	}

	// Response
	cookie.SetRefresh(w, sess.GetRefreshToken(), h.secureCookie)
	returnTo := authHttp.SafeReturnPath(cookie.OAuthReturn(r))
	origin := strings.TrimRight(strings.TrimSpace(h.webOrigin), "/")
	http.Redirect(w, r, origin+returnTo, http.StatusFound)
}
