// Package authPb maps auth domain models to user v1 protobuf messages.
package authPb

import (
	"encoding/json"
	"net"
	"strings"
	"time"

	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Meta builds client hints (device, platform, local network location) from gRPC request fields.
func Meta(userAgent, clientIP string) models.ClientMeta {
	ua := strings.TrimSpace(userAgent)
	ip := strings.TrimSpace(clientIP)
	platform, device := parseUserAgent(ua)
	return models.ClientMeta{
		UserAgent:     ua,
		ClientIP:      ip,
		Platform:      platform,
		DeviceLabel:   device,
		LocationLabel: locationFromIP(ip),
	}
}

func parseUserAgent(ua string) (platform, device string) {
	if ua == "" {
		return "web", ""
	}
	lower := strings.ToLower(ua)

	osName := "Unknown OS"
	switch {
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad"):
		osName = "iOS"
		platform = "mobile"
	case strings.Contains(lower, "android"):
		osName = "Android"
		platform = "mobile"
	case strings.Contains(lower, "mac os x") || strings.Contains(lower, "macintosh"):
		osName = "macOS"
		platform = "web"
	case strings.Contains(lower, "windows"):
		osName = "Windows"
		platform = "web"
	case strings.Contains(lower, "linux"):
		osName = "Linux"
		platform = "web"
	default:
		platform = "web"
	}

	if strings.Contains(lower, "electron") || strings.Contains(lower, "wichatdesktop") {
		platform = "desktop"
	}

	browser := ""
	switch {
	case strings.Contains(lower, "edg/"):
		browser = "Edge"
	case strings.Contains(lower, "opr/") || strings.Contains(lower, "opera"):
		browser = "Opera"
	case strings.Contains(lower, "chrome/") && !strings.Contains(lower, "chromium"):
		browser = "Chrome"
	case strings.Contains(lower, "firefox/"):
		browser = "Firefox"
	case strings.Contains(lower, "safari/") && !strings.Contains(lower, "chrome/"):
		browser = "Safari"
	}

	switch {
	case browser != "" && osName != "Unknown OS":
		return platform, browser + " · " + osName
	case browser != "":
		return platform, browser
	case osName != "Unknown OS":
		return platform, osName
	default:
		return platform, ""
	}
}

func locationFromIP(clientIP string) string {
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return ""
	}
	if ip.IsLoopback() || ip.IsPrivate() {
		return "Local Network"
	}
	return ""
}

// Challenge maps an email OTP or MFA challenge to protobuf.
func Challenge(ch *models.AuthChallenge) *userv1.AuthChallenge {
	if ch == nil {
		return nil
	}

	return &userv1.AuthChallenge{
		ChallengeId:           ch.ChallengeID,
		EmailHint:             ch.EmailHint,
		ResendCooldownSeconds: ch.ResendCooldownSeconds,
		OtpRequired:           !ch.MFARequired,
		MfaRequired:           ch.MFARequired,
		MfaMethods:            ch.MFAMethods,
		WebauthnOptionsJson:   ch.WebAuthnOptionsJSON,
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
	createdAt := ""
	if !user.CreatedAt.IsZero() {
		createdAt = user.CreatedAt.UTC().Format(time.RFC3339Nano)
	}

	statusNoteExpiresAt := ""
	if user.StatusNoteExpiresAt != nil && !user.StatusNoteExpiresAt.IsZero() {
		statusNoteExpiresAt = user.StatusNoteExpiresAt.UTC().Format(time.RFC3339Nano)
	}

	return &userv1.User{
		Id:                  user.ID,
		Email:               user.Email,
		DisplayName:         user.DisplayName,
		Username:            user.Username,
		AvatarUrl:           user.AvatarURL,
		Pronouns:            user.Pronouns,
		Bio:                 user.Bio,
		BannerColor:         user.BannerColor,
		CreatedAt:           createdAt,
		HasPassword:         user.HasPassword,
		StatusNoteText:      user.StatusNoteText,
		StatusNoteEmoji:     user.StatusNoteEmoji,
		StatusNoteExpiresAt: statusNoteExpiresAt,
		Locale:              user.Locale,
		TimeFormat:          user.TimeFormat,
	}
}

// GroupSummary maps a public group to protobuf.
func GroupSummary(g models.PublicGroup) *userv1.GroupSummary {
	return &userv1.GroupSummary{
		Id:                   g.ID,
		Name:                 g.Name,
		LogoUrl:              g.LogoURL,
		DefaultChannelId:     g.DefaultChannelID,
		Role:                 g.Role,
		Description:          g.Description,
		IsPrivate:            g.IsPrivate,
		BannerColor:          g.BannerColor,
		BannerUrl:            g.BannerURL,
		EffectivePermissions: g.EffectivePermissions,
		HideRoles:            g.HideRoles,
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
		CategoryId:     c.CategoryID,
		Position:       c.Position,
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

// ChannelCategory maps a public channel category to protobuf.
func ChannelCategory(c models.PublicChannelCategory) *userv1.ChannelCategory {
	return &userv1.ChannelCategory{
		Id:        c.ID,
		Name:      c.Name,
		IsPrivate: c.IsPrivate,
		Position:  c.Position,
	}
}

// ChannelCategories maps a slice of public channel categories.
func ChannelCategories(categories []models.PublicChannelCategory) []*userv1.ChannelCategory {
	out := make([]*userv1.ChannelCategory, len(categories))
	for i, c := range categories {
		out[i] = ChannelCategory(c)
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
		roles := make([]*userv1.GroupMemberRoleSummary, len(m.Roles))
		for j, r := range m.Roles {
			roles[j] = &userv1.GroupMemberRoleSummary{
				Id:       r.ID,
				Slug:     r.Slug,
				Name:     r.Name,
				Color:    r.Color,
				Hoist:    r.Hoist,
				Position: r.Position,
			}
		}
		presenceStatus := m.PresenceStatus
		if presenceStatus == "" {
			if m.IsOnline {
				presenceStatus = models.PresenceStatusOnline
			} else {
				presenceStatus = models.PresenceStatusOffline
			}
		}
		userCreatedAt := ""
		if !m.UserCreatedAt.IsZero() {
			userCreatedAt = m.UserCreatedAt.UTC().Format(time.RFC3339Nano)
		}
		statusNoteExpiresAt := ""
		if m.StatusNoteExpiresAt != nil && !m.StatusNoteExpiresAt.IsZero() {
			statusNoteExpiresAt = m.StatusNoteExpiresAt.UTC().Format(time.RFC3339Nano)
		}
		timeoutUntil := ""
		if m.TimeoutUntil != nil && !m.TimeoutUntil.IsZero() {
			timeoutUntil = m.TimeoutUntil.UTC().Format(time.RFC3339Nano)
		}
		out[i] = &userv1.GroupMemberEntry{
			UserId:               m.UserID,
			DisplayName:          m.DisplayName,
			Username:             m.Username,
			AvatarUrl:            m.AvatarURL,
			Role:                 m.Role,
			JoinedAt:             m.JoinedAt.UTC().Format(time.RFC3339Nano),
			Roles:                roles,
			IsOnline:             m.IsOnline,
			PresenceStatus:       presenceStatus,
			BannerColor:          m.BannerColor,
			Bio:                  m.Bio,
			Pronouns:             m.Pronouns,
			UserCreatedAt:        userCreatedAt,
			StatusNoteText:       m.StatusNoteText,
			StatusNoteEmoji:      m.StatusNoteEmoji,
			StatusNoteExpiresAt:  statusNoteExpiresAt,
			JoinMethod:           m.JoinMethod,
			InviteCode:           m.InviteCode,
			InvitedByUserId:      m.InvitedByUserID,
			InvitedByUsername:    m.InvitedByUsername,
			InvitedByDisplayName: m.InvitedByDisplayName,
			TimeoutUntil:         timeoutUntil,
		}
	}

	return out
}

// GroupRolePb maps a models.GroupRole to protobuf.
func GroupRolePb(r models.GroupRole) *userv1.GroupRole {
	return &userv1.GroupRole{
		Id:          r.PublicID,
		Slug:        r.Slug,
		Name:        r.Name,
		Color:       r.Color,
		Hoist:       r.Hoist,
		Mentionable: r.Mentionable,
		IsDefault:   r.IsDefault,
		IsSystem:    r.IsSystem,
		Position:    int32(r.Position), //nolint:gosec // G115
		Permissions: r.Permissions,
		MemberCount: int32(r.MemberCount), //nolint:gosec // G115
		CreatedAt:   r.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:   r.UpdatedAt.UTC().Format(time.RFC3339Nano),
	}
}

// GroupRolesPb maps a slice of roles to protobuf.
func GroupRolesPb(roles []models.GroupRole) []*userv1.GroupRole {
	out := make([]*userv1.GroupRole, len(roles))
	for i, r := range roles {
		out[i] = GroupRolePb(r)
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
			Username:    u.Username,
			AvatarUrl:   u.AvatarURL,
		}
	}

	return out
}

// SecurityKey maps a WebAuthn credential model to protobuf.
func SecurityKey(k models.WebAuthnCredential) *userv1.SecurityKey {
	var lastUsed string
	if k.LastUsedAt != nil && !k.LastUsedAt.IsZero() {
		lastUsed = k.LastUsedAt.UTC().Format(time.RFC3339Nano)
	}

	return &userv1.SecurityKey{
		Id:         k.ID,
		Name:       k.Name,
		CreatedAt:  k.CreatedAt.UTC().Format(time.RFC3339Nano),
		LastUsedAt: lastUsed,
		Transports: k.Transports,
	}
}

// SecurityKeys maps a slice of WebAuthn credentials to protobuf.
func SecurityKeys(keys []models.WebAuthnCredential) []*userv1.SecurityKey {
	out := make([]*userv1.SecurityKey, len(keys))
	for i, k := range keys {
		out[i] = SecurityKey(k)
	}
	return out
}

// UserActivityLogs maps domain user activity logs to protobuf.
func UserActivityLogs(logs []models.UserActivityLog) []*userv1.UserActivityLog {
	out := make([]*userv1.UserActivityLog, len(logs))
	for i, entry := range logs {
		changesBytes, _ := json.Marshal(entry.Changes)
		if len(changesBytes) == 0 || string(changesBytes) == "null" {
			changesBytes = []byte("{}")
		}
		metaBytes, _ := json.Marshal(entry.Metadata)
		if len(metaBytes) == 0 || string(metaBytes) == "null" {
			metaBytes = []byte("{}")
		}
		out[i] = &userv1.UserActivityLog{
			Id:           entry.ID,
			UserId:       entry.UserID,
			ActorId:      entry.ActorID,
			Category:     entry.Category,
			Action:       entry.Action,
			TargetType:   entry.TargetType,
			TargetId:     entry.TargetID,
			Summary:      entry.Summary,
			ChangesJson:  string(changesBytes),
			MetadataJson: string(metaBytes),
			CreatedAt:    entry.CreatedAt.UTC().Format(time.RFC3339Nano),
		}
	}
	return out
}
