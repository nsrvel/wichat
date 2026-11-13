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
	Setup(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	VerifyRegister(w http.ResponseWriter, r *http.Request)
	ResendRegisterCode(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	VerifyLogin(w http.ResponseWriter, r *http.Request)
	ResendLoginCode(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Me(w http.ResponseWriter, r *http.Request)
	GoogleStart(w http.ResponseWriter, r *http.Request)
	GoogleCallback(w http.ResponseWriter, r *http.Request)
	ForgotPasswordRequest(w http.ResponseWriter, r *http.Request)
	ForgotPasswordVerify(w http.ResponseWriter, r *http.Request)
	ForgotPasswordReset(w http.ResponseWriter, r *http.Request)
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

func (h handler) Setup(w http.ResponseWriter, r *http.Request) {
	h.Register(w, r)
}

func (h handler) Register(w http.ResponseWriter, r *http.Request) {
	// Bind
	var body authHttp.RegisterBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.Auth.Register(r.Context(), &userv1.RegisterRequest{
		Email:       body.Email,
		Username:    body.Username,
		Password:    body.Password,
		DisplayName: body.DisplayName,
		BirthDate:   body.BirthDate,
		Locale:      body.Locale,
		UserAgent:   middleware.UserAgent(r),
		ClientIp:    middleware.ClientIP(r),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	authHttp.WritePasswordAuth(w, out.GetSession(), out.GetChallenge(), h.secureCookie)
}

func (h handler) VerifyRegister(w http.ResponseWriter, r *http.Request) {
	var body authHttp.VerifyEmailCodeBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.Auth.VerifyRegister(r.Context(), &userv1.VerifyRegisterRequest{
		ChallengeId: body.ChallengeID,
		Code:        body.Code,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	authHttp.WriteSession(w, out.GetSession(), h.secureCookie)
}

func (h handler) ResendRegisterCode(w http.ResponseWriter, r *http.Request) {
	var body authHttp.ResendEmailCodeBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.Auth.ResendRegisterCode(r.Context(), &userv1.ResendRegisterCodeRequest{
		ChallengeId: body.ChallengeID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.ResendEmailCodeData{
		ResendAfterSeconds: out.GetResendCooldownSeconds(),
	}))
}

func (h handler) Login(w http.ResponseWriter, r *http.Request) {
	// Bind
	var body authHttp.LoginBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	login := strings.TrimSpace(body.Login)
	if login == "" {
		login = strings.TrimSpace(body.Email)
	}

	// Service
	out, err := h.ms.User.GRPC.Auth.Login(r.Context(), &userv1.LoginRequest{
		Login:     login,
		Password:  body.Password,
		Locale:    body.Locale,
		UserAgent: middleware.UserAgent(r),
		ClientIp:  middleware.ClientIP(r),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	authHttp.WritePasswordAuth(w, out.GetSession(), out.GetChallenge(), h.secureCookie)
}

func (h handler) VerifyLogin(w http.ResponseWriter, r *http.Request) {
	var body authHttp.VerifyEmailCodeBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.Auth.VerifyLogin(r.Context(), &userv1.VerifyLoginRequest{
		ChallengeId: body.ChallengeID,
		Code:        body.Code,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	authHttp.WriteSession(w, out.GetSession(), h.secureCookie)
}

func (h handler) ResendLoginCode(w http.ResponseWriter, r *http.Request) {
	var body authHttp.ResendEmailCodeBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.Auth.ResendLoginCode(r.Context(), &userv1.ResendLoginCodeRequest{
		ChallengeId: body.ChallengeID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.ResendEmailCodeData{
		ResendAfterSeconds: out.GetResendCooldownSeconds(),
	}))
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
	returnTo := authHttp.SafeReturnPath(cookie.OAuthReturn(r))
	origin := strings.TrimRight(strings.TrimSpace(h.webOrigin), "/")
	http.Redirect(w, r, origin+returnTo, http.StatusFound)
}

func (h handler) ForgotPasswordRequest(w http.ResponseWriter, r *http.Request) {
	// Bind
	var body authHttp.ForgotPasswordRequestBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.Auth.RequestPasswordReset(r.Context(), &userv1.RequestPasswordResetRequest{
		Email:  body.Email,
		Locale: body.Locale,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, map[string]any{
		"data": authHttp.ForgotPasswordRequestData{
			OK:                    true,
			ResendCooldownSeconds: int(out.GetResendCooldownSeconds()),
		},
	})
}

func (h handler) ForgotPasswordVerify(w http.ResponseWriter, r *http.Request) {
	// Bind
	var body authHttp.ForgotPasswordVerifyBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	out, err := h.ms.User.GRPC.Auth.VerifyPasswordReset(r.Context(), &userv1.VerifyPasswordResetRequest{
		Email: body.Email,
		Code:  body.Code,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, map[string]any{
		"data": authHttp.ForgotPasswordVerifyData{ResetToken: out.GetResetToken()},
	})
}

func (h handler) ForgotPasswordReset(w http.ResponseWriter, r *http.Request) {
	// Bind
	var body authHttp.ForgotPasswordResetBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	_, err := h.ms.User.GRPC.Auth.CompletePasswordReset(r.Context(), &userv1.CompletePasswordResetRequest{
		ResetToken: body.ResetToken,
		Password:   body.Password,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, map[string]any{"data": authHttp.OKBody{OK: true}})
}
