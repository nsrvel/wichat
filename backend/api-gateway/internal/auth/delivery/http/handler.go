package http

import (
	"net/http"
	"strings"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/api-gateway/pkg/cookie"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

// Handler is the public auth HTTP delivery.
type Handler interface {
	SetupStatus(w http.ResponseWriter, r *http.Request)
	Setup(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Me(w http.ResponseWriter, r *http.Request)
	GoogleStart(w http.ResponseWriter, r *http.Request)
	GoogleCallback(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	ms           *microservices.Clients
	log          logger.Logger
	secureCookie bool
	webOrigin    string
}

// NewHandler wires auth HTTP to the ms-user client.
func NewHandler(ms *microservices.Clients, log logger.Logger, secureCookie bool, webOrigin string) Handler {
	return handler{
		ms:           ms,
		log:          log,
		secureCookie: secureCookie,
		webOrigin:    strings.TrimRight(strings.TrimSpace(webOrigin), "/"),
	}
}

func (h handler) SetupStatus(w http.ResponseWriter, r *http.Request) {
	// Service
	out, err := h.ms.User.GRPC.Auth.SetupStatus(r.Context(), &userv1.SetupStatusRequest{})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.SetupStatusBody{NeedsSetup: out.GetNeedsSetup()}))
}

func (h handler) Setup(w http.ResponseWriter, r *http.Request) {
	// Bind
	var body authHttp.SetupBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.Auth.Setup(r.Context(), &userv1.SetupRequest{
		Email:       body.Email,
		Password:    body.Password,
		DisplayName: body.DisplayName,
		UserAgent:   middleware.UserAgent(r),
		ClientIp:    middleware.ClientIP(r),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	authHttp.WriteSession(w, out.GetSession(), h.secureCookie)
}

func (h handler) Login(w http.ResponseWriter, r *http.Request) {
	// Bind
	var body authHttp.LoginBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.Auth.Login(r.Context(), &userv1.LoginRequest{
		Email:     body.Email,
		Password:  body.Password,
		UserAgent: middleware.UserAgent(r),
		ClientIp:  middleware.ClientIP(r),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	authHttp.WriteSession(w, out.GetSession(), h.secureCookie)
}

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

func (h handler) GoogleStart(w http.ResponseWriter, r *http.Request) {
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
		authHttp.RedirectSignIn(w, r, h.webOrigin, authHttp.GRPCMessageCode(err))
		return
	}
	sess := out.GetSession()
	if sess == nil || sess.GetRefreshToken() == "" {
		authHttp.RedirectSignIn(w, r, h.webOrigin, string(exception.MsgInternal))
		return
	}

	// Response
	cookie.SetRefresh(w, sess.GetRefreshToken(), h.secureCookie)
	http.Redirect(w, r, h.webOrigin+"/", http.StatusFound)
}
