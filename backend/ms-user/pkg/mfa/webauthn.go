package mfa

import (
	"errors"
	"net/url"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

// WebAuthnUser adapts a WiChat user + credentials to the go-webauthn User interface.
type WebAuthnUser struct {
	ID          []byte
	Name        string
	DisplayName string
	Credentials []webauthn.Credential
}

func (u WebAuthnUser) WebAuthnID() []byte                         { return u.ID }
func (u WebAuthnUser) WebAuthnName() string                       { return u.Name }
func (u WebAuthnUser) WebAuthnDisplayName() string                { return u.DisplayName }
func (u WebAuthnUser) WebAuthnCredentials() []webauthn.Credential { return u.Credentials }

// BuildWebAuthnConfig derives RPID, RPOrigins, and RPDisplayName from explicit config or WebOrigin.
func BuildWebAuthnConfig(webOrigin, rpID, rpOrigins, displayName string) (string, []string, string, error) {
	name := strings.TrimSpace(displayName)
	if name == "" {
		name = "WiChat"
	}

	var origins []string
	for _, part := range strings.Split(rpOrigins, ",") {
		if o := strings.TrimRight(strings.TrimSpace(part), "/"); o != "" {
			origins = append(origins, o)
		}
	}
	if len(origins) == 0 {
		for _, part := range strings.Split(webOrigin, ",") {
			if o := strings.TrimRight(strings.TrimSpace(part), "/"); o != "" {
				origins = append(origins, o)
			}
		}
	}
	if len(origins) == 0 {
		return "", nil, "", errors.New("at least one webauthn origin is required")
	}

	id := strings.TrimSpace(rpID)
	if id == "" {
		parsed, err := url.Parse(origins[0])
		if err != nil || parsed.Hostname() == "" {
			return "", nil, "", errors.New("cannot derive webauthn rp id from origin")
		}
		id = parsed.Hostname()
	}

	return id, origins, name, nil
}

// NewWebAuthn constructs a configured go-webauthn instance.
func NewWebAuthn(webOrigin, rpID, rpOrigins, displayName string) (*webauthn.WebAuthn, error) {
	id, origins, name, err := BuildWebAuthnConfig(webOrigin, rpID, rpOrigins, displayName)
	if err != nil {
		return nil, err
	}

	return webauthn.New(&webauthn.Config{
		RPID:          id,
		RPDisplayName: name,
		RPOrigins:     origins,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			UserVerification: protocol.VerificationPreferred,
			ResidentKey:      protocol.ResidentKeyRequirementPreferred,
		},
	})
}

// CredentialExcludeList converts stored credentials to WebAuthn exclusion descriptors.
func CredentialExcludeList(creds []webauthn.Credential) []protocol.CredentialDescriptor {
	out := make([]protocol.CredentialDescriptor, 0, len(creds))
	for _, c := range creds {
		out = append(out, c.Descriptor())
	}

	return out
}
