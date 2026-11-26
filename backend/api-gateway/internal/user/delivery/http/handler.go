package http

import (
	"net/http"
	"strings"

	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/microservices"
)

// Handler is the public user & instance-setup HTTP delivery.
type Handler interface {
	SetupStatus(w http.ResponseWriter, r *http.Request)
	InstanceSetup(w http.ResponseWriter, r *http.Request)
	CheckUsername(w http.ResponseWriter, r *http.Request)
	SearchUsers(w http.ResponseWriter, r *http.Request)
	UploadAvatar(w http.ResponseWriter, r *http.Request)
	UpdateProfile(w http.ResponseWriter, r *http.Request)
	UpdateStatusNote(w http.ResponseWriter, r *http.Request)
	UpdatePreferences(w http.ResponseWriter, r *http.Request)
	UpdateUsername(w http.ResponseWriter, r *http.Request)
	RequestEmailChange(w http.ResponseWriter, r *http.Request)
	VerifyEmailChange(w http.ResponseWriter, r *http.Request)
	CompleteEmailChange(w http.ResponseWriter, r *http.Request)
	UpdatePassword(w http.ResponseWriter, r *http.Request)
	GetMFAStatus(w http.ResponseWriter, r *http.Request)
	BeginTOTPSetup(w http.ResponseWriter, r *http.Request)
	ConfirmTOTPSetup(w http.ResponseWriter, r *http.Request)
	DisableTOTP(w http.ResponseWriter, r *http.Request)
	BeginWebAuthnRegister(w http.ResponseWriter, r *http.Request)
	FinishWebAuthnRegister(w http.ResponseWriter, r *http.Request)
	RenameWebAuthnCredential(w http.ResponseWriter, r *http.Request)
	DeleteWebAuthnCredential(w http.ResponseWriter, r *http.Request)
	RegenerateRecoveryCodes(w http.ResponseWriter, r *http.Request)
	DisableAccount(w http.ResponseWriter, r *http.Request)
	DeleteAccount(w http.ResponseWriter, r *http.Request)
	ListDevices(w http.ResponseWriter, r *http.Request)
	RevokeDevice(w http.ResponseWriter, r *http.Request)
	RevokeOtherDevices(w http.ResponseWriter, r *http.Request)
	ListActivityLogs(w http.ResponseWriter, r *http.Request)
	GetMyPresence(w http.ResponseWriter, r *http.Request)
	UpdateMyPresence(w http.ResponseWriter, r *http.Request)
	ListRelationships(w http.ResponseWriter, r *http.Request)
	SetRelationship(w http.ResponseWriter, r *http.Request)
	WithPresenceBroadcaster(b PresenceBroadcaster) Handler
	WithTurnstileRequired(required bool) Handler
	WithInviteOnly(inviteOnly bool) Handler
}

// PresenceBroadcaster pushes presence.updated events to subscribed groups.
type PresenceBroadcaster interface {
	DeliverGroup(groupID string, payload []byte) error
}

type handler struct {
	ms                *microservices.Clients
	log               logger.Logger
	turnstileSiteKey  string
	turnstileRequired bool
	inviteOnly        bool
	broadcaster       PresenceBroadcaster
}

// NewHandler wires user HTTP to the ms-user client.
func NewHandler(ms *microservices.Clients, log logger.Logger, turnstileSiteKey ...string) Handler {
	siteKey := ""
	if len(turnstileSiteKey) > 0 {
		siteKey = strings.TrimSpace(turnstileSiteKey[0])
	}
	return handler{ms: ms, log: log, turnstileSiteKey: siteKey}
}

func (h handler) WithTurnstileRequired(required bool) Handler {
	h.turnstileRequired = required
	return h
}

func (h handler) WithInviteOnly(inviteOnly bool) Handler {
	h.inviteOnly = inviteOnly
	return h
}

func (h handler) WithPresenceBroadcaster(b PresenceBroadcaster) Handler {
	h.broadcaster = b
	return h
}
