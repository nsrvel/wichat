package http

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
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
	InstanceSetup(w http.ResponseWriter, r *http.Request)
	ListMyGroups(w http.ResponseWriter, r *http.Request)
	ReorderMyGroups(w http.ResponseWriter, r *http.Request)
	ListJoinableGroups(w http.ResponseWriter, r *http.Request)
	CreateGroup(w http.ResponseWriter, r *http.Request)
	JoinGroup(w http.ResponseWriter, r *http.Request)
	LeaveGroup(w http.ResponseWriter, r *http.Request)
	DeleteGroup(w http.ResponseWriter, r *http.Request)
	ListGroupChannels(w http.ResponseWriter, r *http.Request)
	UploadGroupIcon(w http.ResponseWriter, r *http.Request)
	UpdateGroupProfile(w http.ResponseWriter, r *http.Request)
	UploadGroupBanner(w http.ResponseWriter, r *http.Request)
	ListGroupMembers(w http.ResponseWriter, r *http.Request)
	SearchUsers(w http.ResponseWriter, r *http.Request)
	GetGroupInviteLink(w http.ResponseWriter, r *http.Request)
	SendGroupInvites(w http.ResponseWriter, r *http.Request)
	PreviewGroupInvite(w http.ResponseWriter, r *http.Request)
	RedeemGroupInvite(w http.ResponseWriter, r *http.Request)
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
	out, err := h.ms.User.GRPC.User.SetupStatus(r.Context(), &userv1.SetupStatusRequest{})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.SetupStatusBody{
		RegistrationMode: out.GetRegistrationMode(),
		UsernameEnabled:  out.GetUsernameEnabled(),
	}))
}

func (h handler) InstanceSetup(w http.ResponseWriter, r *http.Request) {
	// Bind
	token := authHttp.Bearer(r)
	if token == "" {
		responsehttp.WriteError(w, r, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}
	var body authHttp.InstanceSetupBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Service
	_, err := h.ms.User.GRPC.User.CompleteInstanceSetup(r.Context(), &userv1.CompleteInstanceSetupRequest{
		AccessToken:      token,
		Name:             body.Name,
		LogoUrl:          body.LogoURL,
		UsernameEnabled:  body.UsernameEnabled,
		RegistrationMode: body.RegistrationMode,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	// Response
	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
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

func (h handler) ListMyGroups(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	out, err := h.ms.User.GRPC.Group.ListMyGroups(r.Context(), &userv1.ListMyGroupsRequest{AccessToken: token})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupsListBody{
		Groups: authHttp.GroupsFromPB(out.GetGroups()),
	}))
}

func (h handler) ReorderMyGroups(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	var body authHttp.ReorderMyGroupsBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}
	if len(body.GroupIDs) == 0 {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "group_ids"}))
		return
	}

	_, err := h.ms.User.GRPC.Group.ReorderMyGroups(r.Context(), &userv1.ReorderMyGroupsRequest{
		AccessToken: token,
		GroupIds:    body.GroupIDs,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, map[string]any{"data": authHttp.OKBody{OK: true}})
}

func (h handler) ListJoinableGroups(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	out, err := h.ms.User.GRPC.Group.ListJoinableGroups(r.Context(), &userv1.ListJoinableGroupsRequest{AccessToken: token})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupsListBody{
		Groups: authHttp.GroupsFromPB(out.GetGroups()),
	}))
}

func (h handler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	var body authHttp.CreateGroupBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	name := strings.TrimSpace(body.Name)
	if name == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "name"}))
		return
	}

	out, err := h.ms.User.GRPC.Group.CreateGroup(r.Context(), &userv1.CreateGroupRequest{
		AccessToken: token,
		Name:        name,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusCreated, response.NewDataEnvelope(authHttp.GroupFromPB(out.GetGroup())))
}

func (h handler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("id"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "id"}))
		return
	}

	_, err := h.ms.User.GRPC.Group.JoinGroup(r.Context(), &userv1.JoinGroupRequest{
		AccessToken: token,
		GroupId:     groupID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("id"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "id"}))
		return
	}

	_, err := h.ms.User.GRPC.Group.LeaveGroup(r.Context(), &userv1.LeaveGroupRequest{
		AccessToken: token,
		GroupId:     groupID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("id"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "id"}))
		return
	}

	_, err := h.ms.User.GRPC.Group.DeleteGroup(r.Context(), &userv1.DeleteGroupRequest{
		AccessToken: token,
		GroupId:     groupID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.OKBody{OK: true}))
}

func (h handler) ListGroupChannels(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("id"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "id"}))
		return
	}

	out, err := h.ms.User.GRPC.Channel.ListGroupChannels(r.Context(), &userv1.ListGroupChannelsRequest{
		AccessToken: token,
		GroupId:     groupID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.ChannelsListBody{
		Channels: authHttp.ChannelsFromPB(out.GetChannels()),
	}))
}

func (h handler) UploadGroupIcon(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("id"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "id"}))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, authHttp.MaxGroupIconBytes+(1<<20))
	if err := r.ParseMultipartForm(authHttp.MaxGroupIconBytes + (1 << 10)); err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, authHttp.MaxGroupIconBytes+1))
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}
	if len(data) == 0 || len(data) > authHttp.MaxGroupIconBytes {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	out, err := h.ms.User.GRPC.Group.UploadGroupIcon(r.Context(), &userv1.UploadGroupIconRequest{
		AccessToken: token,
		GroupId:     groupID,
		ContentType: contentType,
		Data:        data,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupFromPB(out.GetGroup())))
}

func (h handler) UpdateGroupProfile(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("id"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "id"}))
		return
	}

	var body authHttp.UpdateGroupProfileBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		return
	}

	out, err := h.ms.User.GRPC.Group.UpdateGroupProfile(r.Context(), &userv1.UpdateGroupProfileRequest{
		AccessToken:      token,
		GroupId:          groupID,
		Name:             body.Name,
		Description:      body.Description,
		IsPrivate:        body.IsPrivate,
		BannerColor:      body.BannerColor,
		ClearLogo:        body.ClearLogo,
		ClearBannerImage: body.ClearBannerImage,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupFromPB(out.GetGroup())))
}

func (h handler) UploadGroupBanner(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("id"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "id"}))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, authHttp.MaxGroupBannerBytes+(1<<20))
	if err := r.ParseMultipartForm(authHttp.MaxGroupBannerBytes + (1 << 10)); err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, authHttp.MaxGroupBannerBytes+1))
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}
	if len(data) == 0 || len(data) > authHttp.MaxGroupBannerBytes {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"}))
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	out, err := h.ms.User.GRPC.Group.UploadGroupBanner(r.Context(), &userv1.UploadGroupBannerRequest{
		AccessToken: token,
		GroupId:     groupID,
		ContentType: contentType,
		Data:        data,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupFromPB(out.GetGroup())))
}

func (h handler) ListGroupMembers(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("id"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "id"}))
		return
	}

	limit := 20
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}

	out, err := h.ms.User.GRPC.Group.ListGroupMembers(r.Context(), &userv1.ListGroupMembersRequest{
		AccessToken: token,
		GroupId:     groupID,
		Limit:       int32(limit), //nolint:gosec // G115 — query limit clamped above
		Cursor:      strings.TrimSpace(r.URL.Query().Get("cursor")),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupMembersPageBody{
		Members:    authHttp.GroupMembersFromPB(out.GetMembers()),
		NextCursor: out.GetNextCursor(),
		Total:      int(out.GetTotal()),
	}))
}

func (h handler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	groupID := strings.TrimSpace(r.URL.Query().Get("exclude_group_id"))
	limit := 20
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}

	out, err := h.ms.User.GRPC.Group.SearchUsers(r.Context(), &userv1.SearchUsersRequest{
		AccessToken: token,
		Query:       query,
		Limit:       int32(limit), //nolint:gosec // G115 — query limit clamped above
		GroupId:     groupID,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.UserSearchBody{
		Users: authHttp.UserSearchFromPB(out.GetUsers()),
	}))
}

func (h handler) GetGroupInviteLink(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("id"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "id"}))
		return
	}

	out, err := h.ms.User.GRPC.Group.GetGroupInviteLink(r.Context(), &userv1.GetGroupInviteLinkRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   strings.TrimSpace(r.URL.Query().Get("channel_id")),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	inviteURL := h.webOrigin + "/invite/" + url.PathEscape(out.GetToken())

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.GroupInviteLinkBody{
		Token:     out.GetToken(),
		URL:       inviteURL,
		ExpiresAt: out.GetExpiresAt(),
	}))
}

func (h handler) SendGroupInvites(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	groupID := strings.TrimSpace(r.PathValue("id"))
	if groupID == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "id"}))
		return
	}

	var body authHttp.SendGroupInvitesBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.Group.SendGroupInvites(r.Context(), &userv1.SendGroupInvitesRequest{
		AccessToken: token,
		GroupId:     groupID,
		ChannelId:   strings.TrimSpace(body.ChannelID),
		UserIds:     body.UserIDs,
		Emails:      body.Emails,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.SendGroupInvitesResultBody{
		Added:   int(out.GetAdded()),
		Emailed: int(out.GetEmailed()),
		Skipped: int(out.GetSkipped()),
	}))
}

func (h handler) PreviewGroupInvite(w http.ResponseWriter, r *http.Request) {
	inviteToken := strings.TrimSpace(r.URL.Query().Get("invite"))
	if inviteToken == "" {
		inviteToken = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	if inviteToken == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "invite"}))
		return
	}

	out, err := h.ms.User.GRPC.Group.PreviewGroupInvite(r.Context(), &userv1.PreviewGroupInviteRequest{
		Token: inviteToken,
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.InvitePreviewBody{
		Valid:        out.GetValid(),
		GroupName:    out.GetGroupName(),
		ChannelName:  out.GetChannelName(),
		GroupLogoURL: out.GetGroupLogoUrl(),
		MemberCount:  int(out.GetMemberCount()),
		OnlineCount:  int(out.GetOnlineCount()),
		ExpiresAt:    out.GetExpiresAt(),
		Reason:       out.GetReason(),
	}))
}

func (h handler) RedeemGroupInvite(w http.ResponseWriter, r *http.Request) {
	token := authHttp.Bearer(r)
	if token == "" {
		authHttp.WriteOutbound(w, r, h.log, exception.Unauthenticated(exception.MsgUnauthorized, nil))
		return
	}

	var body authHttp.RedeemInviteBody
	if err := authHttp.DecodeJSON(w, r, &body); err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	out, err := h.ms.User.GRPC.Group.RedeemGroupInvite(r.Context(), &userv1.RedeemGroupInviteRequest{
		AccessToken: token,
		Token:       strings.TrimSpace(body.Token),
	})
	if err != nil {
		authHttp.WriteOutbound(w, r, h.log, err)
		return
	}

	responsehttp.WriteJSON(w, http.StatusOK, response.NewDataEnvelope(authHttp.RedeemInviteResultBody{
		GroupID:   out.GetGroupId(),
		ChannelID: out.GetChannelId(),
	}))
}
