// Package authHttp maps auth JSON bodies and protobuf sessions for HTTP responses (dir: auth-http, file: auth_http.go).
package authHttp

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/wichat/wichat/backend/api-gateway/pkg/cookie"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
	"github.com/wichat/wichat/backend/wi-shared/response"
	responsegrpc "github.com/wichat/wichat/backend/wi-shared/response/grpc"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

const maxBodyBytes = 1 << 20

// SetupBody is the deprecated JSON body for POST /api/v1/auth/setup.
type SetupBody = RegisterBody

// RegisterBody is the JSON body for POST /api/v1/auth/register.
type RegisterBody struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	BirthDate   string `json:"birth_date"`
	Locale      string `json:"locale"`
}

// LoginBody is the JSON body for POST /api/v1/auth/login.
type LoginBody struct {
	Login    string `json:"login"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Locale   string `json:"locale"`
}

// AuthChallengeBody is the email OTP step before a session is issued.
type AuthChallengeBody struct {
	OTPRequired        bool   `json:"otp_required"`
	ChallengeID        string `json:"challenge_id"`
	EmailHint          string `json:"email_hint"`
	ResendAfterSeconds int32  `json:"resend_after_seconds"`
}

// VerifyEmailCodeBody is POST /api/v1/auth/register/verify and login/verify.
type VerifyEmailCodeBody struct {
	ChallengeID string `json:"challenge_id"`
	Code        string `json:"code"`
}

// ResendEmailCodeBody is POST resend endpoints.
type ResendEmailCodeBody struct {
	ChallengeID string `json:"challenge_id"`
}

// ResendEmailCodeData is resend success payload.
type ResendEmailCodeData struct {
	ResendAfterSeconds int32 `json:"resend_after_seconds"`
}

// GuestRegisterBody is the JSON body for POST /api/v1/auth/register/guest.
type GuestRegisterBody struct {
	DisplayName string `json:"display_name"`
	BirthDate   string `json:"birth_date"`
}

// SetupStatusBody is the JSON data for GET /api/v1/auth/setup-status (instance policy).
type SetupStatusBody struct {
	RegistrationMode string `json:"registration_mode"`
	GuestEnabled     bool   `json:"guest_enabled"`
	UsernameEnabled  bool   `json:"username_enabled"`
}

// InstanceSetupBody is POST /api/v1/auth/instance-setup.
type InstanceSetupBody struct {
	Name             string `json:"name"`
	LogoURL          string `json:"logo_url"`
	UsernameEnabled  bool   `json:"username_enabled"`
	RegistrationMode string `json:"registration_mode"`
	GuestEnabled     bool   `json:"guest_enabled"`
}

// SessionBody is the public session JSON (refresh token is HttpOnly cookie only).
type SessionBody struct {
	AccessToken string   `json:"access_token"`
	ExpiresIn   int32    `json:"expires_in"`
	User        UserBody `json:"user"`
}

// UserBody is the public user slice in auth responses.
type UserBody struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	IsGuest     bool   `json:"is_guest"`
}

// OKBody is a minimal success payload.
type OKBody struct {
	OK bool `json:"ok"`
}

// GroupBody is a public group in JSON responses.
type GroupBody struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	LogoURL string `json:"logo_url"`
}

// GroupsListBody wraps a group list.
type GroupsListBody struct {
	Groups []GroupBody `json:"groups"`
}

// CreateGroupBody is POST /api/v1/groups.
type CreateGroupBody struct {
	Name string `json:"name"`
}

// ForgotPasswordRequestData is POST /api/v1/auth/forgot-password success body.
type ForgotPasswordRequestData struct {
	OK                    bool `json:"ok"`
	ResendCooldownSeconds int  `json:"resend_cooldown_seconds"`
}

// ForgotPasswordRequestBody is POST /api/v1/auth/forgot-password.
type ForgotPasswordRequestBody struct {
	Email  string `json:"email"`
	Locale string `json:"locale"`
}

// ForgotPasswordVerifyBody is POST /api/v1/auth/forgot-password/verify.
type ForgotPasswordVerifyBody struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// ForgotPasswordResetBody is POST /api/v1/auth/forgot-password/reset.
type ForgotPasswordResetBody struct {
	ResetToken string `json:"reset_token"`
	Password   string `json:"password"`
}

// ForgotPasswordVerifyData is the verify step response data.
type ForgotPasswordVerifyData struct {
	ResetToken string `json:"reset_token"`
}

// DecodeJSON reads a JSON body with a size limit.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return exception.InvalidInput(exception.MsgInvalidInput, nil)
	}

	return nil
}

// Bearer returns the token from Authorization: Bearer.
func Bearer(r *http.Request) string {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return ""
	}

	return strings.TrimSpace(header[len(prefix):])
}

// GroupFromPB maps a group summary proto to JSON.
func GroupFromPB(g *userv1.GroupSummary) GroupBody {
	if g == nil {
		return GroupBody{}
	}

	return GroupBody{
		ID:      g.GetId(),
		Name:    g.GetName(),
		LogoURL: g.GetLogoUrl(),
	}
}

// GroupsFromPB maps group summary protos to JSON bodies.
func GroupsFromPB(groups []*userv1.GroupSummary) []GroupBody {
	out := make([]GroupBody, len(groups))
	for i, g := range groups {
		out[i] = GroupFromPB(g)
	}

	return out
}

// UserFromPB maps auth user proto to JSON body.
func UserFromPB(user *userv1.User) UserBody {
	return UserBody{
		ID:          user.GetId(),
		Email:       user.GetEmail(),
		DisplayName: user.GetDisplayName(),
		Username:    user.GetUsername(),
		IsGuest:     user.GetIsGuest(),
	}
}

// ChallengeFromPB maps an auth challenge proto to JSON.
func ChallengeFromPB(ch *userv1.AuthChallenge) AuthChallengeBody {
	if ch == nil {
		return AuthChallengeBody{}
	}

	return AuthChallengeBody{
		OTPRequired:        ch.GetOtpRequired(),
		ChallengeID:        ch.GetChallengeId(),
		EmailHint:          ch.GetEmailHint(),
		ResendAfterSeconds: ch.GetResendCooldownSeconds(),
	}
}

// WriteChallenge returns the email OTP step without setting a session cookie.
func WriteChallenge(w http.ResponseWriter, ch *userv1.AuthChallenge) {
	body := ChallengeFromPB(ch)
	if body.ChallengeID == "" {
		responsehttp.WriteError(w, nil, exception.Internal(nil))
		return
	}
	body.OTPRequired = true
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(body))
}

// WritePasswordAuth writes either a challenge or a session from register/login RPCs.
func WritePasswordAuth(w http.ResponseWriter, sess *userv1.Session, ch *userv1.AuthChallenge, secure bool) {
	if ch != nil && ch.GetChallengeId() != "" {
		WriteChallenge(w, ch)
		return
	}

	WriteSession(w, sess, secure)
}

// WriteSession sets the refresh cookie and returns access token + user in JSON.
func WriteSession(w http.ResponseWriter, sess *userv1.Session, secure bool) {
	if sess == nil || sess.GetAccessToken() == "" || sess.GetRefreshToken() == "" {
		responsehttp.WriteError(w, nil, exception.Internal(nil))
		return
	}

	cookie.SetRefresh(w, sess.GetRefreshToken(), secure)
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(SessionBody{
		AccessToken: sess.GetAccessToken(),
		ExpiresIn:   sess.GetExpiresIn(),
		User:        UserFromPB(sess.GetUser()),
	}))
}

// WriteOutbound maps downstream gRPC/transport errors to JSON (optional app log on 5xx).
func WriteOutbound(w http.ResponseWriter, r *http.Request, log logger.Logger, err error) {
	var opts []responsehttp.WriteOption
	if log != nil {
		opts = append(opts, responsehttp.WithLogger(logger.HTTPWriteErrorLogger(log)))
	}
	responsehttp.WriteOutboundError(w, r, err, opts...)
}

// GRPCMessageCode extracts an i18n error code from a gRPC status for OAuth redirects.
func GRPCMessageCode(err error) string {
	parsed, ok := responsegrpc.FromError(err)
	if !ok {
		return string(exception.MsgInternal)
	}

	return string(parsed.Code)
}

// RedirectSignIn sends the browser to the web sign-in page with an error query param.
func RedirectSignIn(w http.ResponseWriter, r *http.Request, webOrigin, code string) {
	target := strings.TrimRight(strings.TrimSpace(webOrigin), "/") + "/sign-in?error=" + url.QueryEscape(code)
	http.Redirect(w, r, target, http.StatusFound)
}
