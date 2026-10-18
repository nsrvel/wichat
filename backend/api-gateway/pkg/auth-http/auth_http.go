// Package authHttp maps auth JSON bodies and protobuf sessions for HTTP responses.
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

// SetupBody is the JSON body for POST /api/v1/auth/setup.
type SetupBody struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// LoginBody is the JSON body for POST /api/v1/auth/login.
type LoginBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SetupStatusBody is the JSON data for setup-status.
type SetupStatusBody struct {
	NeedsSetup bool `json:"needs_setup"`
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
}

// OKBody is a minimal success payload.
type OKBody struct {
	OK bool `json:"ok"`
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

// UserFromPB maps auth user proto to JSON body.
func UserFromPB(user *userv1.User) UserBody {
	return UserBody{
		ID:          user.GetId(),
		Email:       user.GetEmail(),
		DisplayName: user.GetDisplayName(),
	}
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
