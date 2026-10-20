// Package authPb maps auth domain models to user v1 protobuf messages.
package authPb

import (
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Meta builds client hints from gRPC request fields.
func Meta(userAgent, clientIP string) models.ClientMeta {
	return models.ClientMeta{UserAgent: userAgent, ClientIP: clientIP}
}

// Challenge maps an email OTP challenge to protobuf.
func Challenge(ch *models.AuthChallenge) *userv1.AuthChallenge {
	if ch == nil {
		return nil
	}

	return &userv1.AuthChallenge{
		ChallengeId:           ch.ChallengeID,
		EmailHint:             ch.EmailHint,
		ResendCooldownSeconds: ch.ResendCooldownSeconds,
		OtpRequired:           true,
	}
}

// PasswordAuth maps a register/login step result to protobuf.
func PasswordAuth(out *models.PasswordAuthResult) (*userv1.Session, *userv1.AuthChallenge) {
	if out == nil {
		return nil, nil
	}
	if out.Session != nil {
		return Session(out.Session), nil
	}

	return nil, Challenge(out.Challenge)
}

// Session maps a service session result to protobuf.
func Session(out *models.SessionResult) *userv1.Session {
	if out == nil {
		return nil
	}

	return &userv1.Session{
		AccessToken:  out.AccessToken,
		ExpiresIn:    out.ExpiresIn,
		RefreshToken: out.RefreshToken,
		User:         User(out.User),
	}
}

// User maps a public user to protobuf.
func User(user models.PublicUser) *userv1.User {
	return &userv1.User{
		Id:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Username:    user.Username,
		IsGuest:     user.IsGuest,
	}
}

// GroupSummary maps a public group to protobuf.
func GroupSummary(g models.PublicGroup) *userv1.GroupSummary {
	return &userv1.GroupSummary{
		Id:      g.ID,
		Name:    g.Name,
		LogoUrl: g.LogoURL,
	}
}

// GroupSummaries maps a slice of public groups.
func GroupSummaries(groups []models.PublicGroup) []*userv1.GroupSummary {
	out := make([]*userv1.GroupSummary, len(groups))
	for i, g := range groups {
		out[i] = GroupSummary(g)
	}

	return out
}
