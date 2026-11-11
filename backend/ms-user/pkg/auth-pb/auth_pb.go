// Package authPb maps auth domain models to user v1 protobuf messages.
package authPb

import (
	"time"

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
	}
}

// GroupSummary maps a public group to protobuf.
func GroupSummary(g models.PublicGroup) *userv1.GroupSummary {
	return &userv1.GroupSummary{
		Id:               g.ID,
		Name:             g.Name,
		LogoUrl:          g.LogoURL,
		DefaultChannelId: g.DefaultChannelID,
		Role:             g.Role,
		Description:      g.Description,
		IsPrivate:        g.IsPrivate,
		BannerColor:      g.BannerColor,
		BannerUrl:        g.BannerURL,
	}
}

// ChannelSummary maps a public channel to protobuf.
func ChannelSummary(c models.PublicChannel) *userv1.ChannelSummary {
	return &userv1.ChannelSummary{
		Id:             c.ID,
		Name:           c.Name,
		Type:           c.Type,
		SidebarSection: c.SidebarSection,
		ImageUrl:       c.ImageURL,
		IsPrivate:      c.IsPrivate,
	}
}

// ChannelSummaries maps a slice of public channels.
func ChannelSummaries(channels []models.PublicChannel) []*userv1.ChannelSummary {
	out := make([]*userv1.ChannelSummary, len(channels))
	for i, c := range channels {
		out[i] = ChannelSummary(c)
	}

	return out
}

// GroupSummaries maps a slice of public groups.
func GroupSummaries(groups []models.PublicGroup) []*userv1.GroupSummary {
	out := make([]*userv1.GroupSummary, len(groups))
	for i, g := range groups {
		out[i] = GroupSummary(g)
	}

	return out
}

// GroupMemberEntries maps member profiles to protobuf.
func GroupMemberEntries(members []models.GroupMemberProfile) []*userv1.GroupMemberEntry {
	out := make([]*userv1.GroupMemberEntry, len(members))
	for i, m := range members {
		out[i] = &userv1.GroupMemberEntry{
			UserId:      m.UserID,
			DisplayName: m.DisplayName,
			Email:       m.Email,
			Role:        m.Role,
			JoinedAt:    m.JoinedAt.UTC().Format(time.RFC3339Nano),
		}
	}

	return out
}

// UserSearchHits maps search rows to protobuf.
func UserSearchHits(users []models.UserSearchHit) []*userv1.UserSearchHit {
	out := make([]*userv1.UserSearchHit, len(users))
	for i, u := range users {
		out[i] = &userv1.UserSearchHit{
			UserId:      u.UserID,
			DisplayName: u.DisplayName,
			Email:       u.Email,
			Username:    u.Username,
		}
	}

	return out
}
