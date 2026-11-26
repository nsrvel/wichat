package http

import (
	"net/http"
	"strings"

	"github.com/wichat/wichat/backend/api-gateway/internal/middleware"
	authHttp "github.com/wichat/wichat/backend/api-gateway/pkg/auth-http"
	"github.com/wichat/wichat/backend/api-gateway/pkg/turnstile"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
)

// Handler is the public auth HTTP delivery.
type Handler interface {
	Setup(w http.ResponseWriter, r *http.Request)
	Register(w http.ResponseWriter, r *http.Request)
	VerifyRegister(w http.ResponseWriter, r *http.Request)
	ResendRegisterCode(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	VerifyLogin(w http.ResponseWriter, r *http.Request)
	VerifyLoginMFA(w http.ResponseWriter, r *http.Request)
	BeginLoginWebAuthn(w http.ResponseWriter, r *http.Request)
	ResendLoginCode(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	Me(w http.ResponseWriter, r *http.Request)
	GoogleStart(w http.ResponseWriter, r *http.Request)
	GoogleCallback(w http.ResponseWriter, r *http.Request)
	ForgotPasswordRequest(w http.ResponseWriter, r *http.Request)
	ForgotPasswordVerify(w http.ResponseWriter, r *http.Request)
	ForgotPasswordReset(w http.ResponseWriter, r *http.Request)
	WithInviteOnly(inviteOnly bool) Handler
}

type handler struct {
	ms           *microservices.Clients
	log          logger.Logger
	secureCookie bool
	webOrigin    string
	turnstile    turnstile.Verifier
	inviteOnly   bool
}

// WithInviteOnly mirrors INVITE_ONLY for OAuth redirect handling on the HTTP edge.
func (h handler) WithInviteOnly(inviteOnly bool) Handler {
	h.inviteOnly = inviteOnly
	return h
}

// NewHandler wires auth HTTP to the ms-user client.
func NewHandler(ms *microservices.Clients, log logger.Logger, secureCookie bool, webOrigin string, turnstileSecret ...string) Handler {
	secret := ""
	if len(turnstileSecret) > 0 {
		secret = turnstileSecret[0]
	}
	return handler{
		ms:           ms,
		log:          log,
		secureCookie: secureCookie,
		webOrigin:    strings.TrimRight(strings.TrimSpace(webOrigin), "/"),
		turnstile:    turnstile.NewVerifier(secret),
	}
}

func (h handler) verifyTurnstile(w http.ResponseWriter, r *http.Request, token string) bool {
	if err := h.turnstile.Verify(r.Context(), token, middleware.ClientIP(r)); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return false
	}
	return true
}
