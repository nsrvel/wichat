package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	channelrepo "github.com/wichat/wichat/backend/ms-user/internal/channel/repository"
	groupactivitysvc "github.com/wichat/wichat/backend/ms-user/internal/group-activity-log/service"
	groupmemberrepo "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository"
	grouprolerepo "github.com/wichat/wichat/backend/ms-user/internal/group-role/repository"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	inviterepo "github.com/wichat/wichat/backend/ms-user/internal/invite/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	presencesvc "github.com/wichat/wichat/backend/ms-user/internal/presence/service"
	activitysvc "github.com/wichat/wichat/backend/ms-user/internal/user-activity-log/service"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	"github.com/wichat/wichat/backend/ms-user/pkg/access"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/objecturl"
	"github.com/wichat/wichat/backend/ms-user/pkg/token"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	notificationv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/notification/v1"
)

// Notifier sends outbound email via ms-notification.
type Notifier interface {
	Send(ctx context.Context, in *notificationv1.SendRequest) (*notificationv1.SendResponse, error)
}

// Service is the group invite and share use-case API.
type Service interface {
	GetInviteLink(ctx context.Context, accessToken, groupPublicID, channelPublicID string) (*models.InviteLinkResult, error)
	SendGroupInvites(ctx context.Context, accessToken string, in models.SendGroupInvitesInput) (*models.SendGroupInvitesResult, error)
	PreviewInvite(ctx context.Context, inviteToken string) (*models.InvitePreviewResult, error)
	RedeemInvite(ctx context.Context, accessToken, inviteToken string) (*models.RedeemInviteResult, error)
	WithActivityLog(r activitysvc.Recorder) Service
	WithGroupActivityLog(r groupactivitysvc.Recorder) Service
	WithPresence(p presencesvc.Reader) Service
}

type service struct {
	groups           grouprepo.PgRepository
	members          groupmemberrepo.PgRepository
	roles            grouprolerepo.PgRepository
	channels         channelrepo.PgRepository
	invites          inviterepo.PgRepository
	users            userrepo.PgRepository
	sessions         authrepo.SessionRepository
	notify           Notifier
	activity         activitysvc.Recorder
	groupActivity    groupactivitysvc.Recorder
	presence         presencesvc.Reader
	secret           []byte
	webBase          string
	objectPublicBase string
	log              logger.Logger
	now              func() time.Time
}

// NewService builds invite use cases.
func NewService(
	groups grouprepo.PgRepository,
	members groupmemberrepo.PgRepository,
	roles grouprolerepo.PgRepository,
	channels channelrepo.PgRepository,
	invites inviterepo.PgRepository,
	users userrepo.PgRepository,
	sessions authrepo.SessionRepository,
	notify Notifier,
	secret []byte,
	webOrigin string,
	objectPublicBase string,
	log logger.Logger,
) Service {
	copied := make([]byte, len(secret))
	copy(copied, secret)

	return service{
		groups:           groups,
		members:          members,
		roles:            roles,
		channels:         channels,
		invites:          invites,
		users:            users,
		sessions:         sessions,
		notify:           notify,
		secret:           copied,
		webBase:          strings.TrimRight(strings.TrimSpace(webOrigin), "/"),
		objectPublicBase: strings.TrimSpace(objectPublicBase),
		log:              log,
		now:              func() time.Time { return time.Now().UTC() },
	}
}

func (s service) WithActivityLog(r activitysvc.Recorder) Service {
	s.activity = r
	return s
}

func (s service) WithGroupActivityLog(r groupactivitysvc.Recorder) Service {
	s.groupActivity = r
	return s
}

func (s service) recordGroupActivity(ctx context.Context, entry models.GroupActivityLogEntry) {
	if s.groupActivity == nil {
		return
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = s.now()
	}
	s.groupActivity.Record(ctx, entry)
}

func (s service) WithPresence(p presencesvc.Reader) Service {
	s.presence = p
	return s
}

func (s service) recordActivity(ctx context.Context, entry models.UserActivityLogEntry) {
	if s.activity == nil {
		return
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = s.now()
	}
	s.activity.Record(ctx, entry)
}

func (s service) GetInviteLink(
	ctx context.Context,
	accessToken, groupPublicID, channelPublicID string,
) (*models.InviteLinkResult, error) {
	group, member, err := s.loadGroupMember(ctx, accessToken, groupPublicID)
	if err != nil {
		return nil, err
	}

	perms, err := s.roles.GetEffectivePermissions(ctx, group.ID, member.UserID)
	if err != nil {
		return nil, err
	}
	if !access.HasPermission(member.Role == constants.GroupRoleOwner, perms, constants.PermCreateInvite) {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	var channelID *string
	if chID, err := s.resolveChannelID(ctx, group.ID, channelPublicID); err != nil {
		return nil, err
	} else if chID != "" {
		channelID = &chID
	}

	now := s.now()
	row, err := s.invites.FindOpenShareLink(ctx, group.ID, member.UserID, channelID, now)
	if err != nil && !errors.Is(err, constants.ErrNotFound) {
		return nil, err
	}
	needNew := errors.Is(err, constants.ErrNotFound)
	if needNew {
		tokenStr, err := token.NewInviteToken()
		if err != nil {
			return nil, err
		}
		id, err := token.NewID()
		if err != nil {
			return nil, err
		}
		userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
		if err != nil {
			return nil, err
		}
		row = models.GroupInviteRow{
			ID:              id,
			Token:           tokenStr,
			GroupID:         group.ID,
			ChannelID:       channelID,
			InvitedByUserID: userID,
			Kind:            constants.GroupInviteKindShareLink,
			ExpiresAt:       now.Add(constants.GroupShareLinkTTL),
			CreatedAt:       now,
		}
		if err := s.invites.InsertInvite(ctx, row); err != nil {
			return nil, err
		}
		s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
			GroupID:    group.ID,
			ActorID:    userID,
			Category:   constants.GroupAuditCategoryInvite,
			Action:     constants.GroupAuditInviteCreated,
			TargetType: constants.GroupAuditTargetInvite,
			TargetID:   row.Token,
			TargetName: row.Token,
			Summary:    "Created invite link",
			Metadata:   map[string]any{"code": row.Token},
			CreatedAt:  now,
		})
	}

	return &models.InviteLinkResult{Token: row.Token, ExpiresAt: row.ExpiresAt}, nil
}

func (s service) SendGroupInvites(
	ctx context.Context,
	accessToken string,
	in models.SendGroupInvitesInput,
) (*models.SendGroupInvitesResult, error) {
	group, member, err := s.loadGroupMember(ctx, accessToken, in.GroupPublicID)
	if err != nil {
		return nil, err
	}

	perms, err := s.roles.GetEffectivePermissions(ctx, group.ID, member.UserID)
	if err != nil {
		return nil, err
	}
	if !access.HasPermission(member.Role == constants.GroupRoleOwner, perms, constants.PermCreateInvite) {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	var channelID *string
	if chID, err := s.resolveChannelID(ctx, group.ID, in.ChannelPublicID); err != nil {
		return nil, err
	} else if chID != "" {
		channelID = &chID
	}

	if len(in.UserIDs)+len(in.Emails) == 0 {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "recipients"})
	}
	if len(in.UserIDs)+len(in.Emails) > constants.MaxInviteBatchSize {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "recipients"})
	}

	now := s.now()
	inviterID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, err
	}
	inviter, err := s.users.FindUserByID(ctx, inviterID)
	if err != nil {
		return nil, err
	}

	out := &models.SendGroupInvitesResult{}
	seenUsers := make(map[string]struct{}, len(in.UserIDs))

	for _, rawID := range in.UserIDs {
		id := strings.TrimSpace(rawID)
		if id == "" {
			out.Skipped++
			continue
		}
		if _, dup := seenUsers[id]; dup {
			out.Skipped++
			continue
		}
		seenUsers[id] = struct{}{}

		if _, err := s.members.FindGroupMember(ctx, group.ID, id); err == nil {
			out.Skipped++
			continue
		} else if !errors.Is(err, constants.ErrNotFound) {
			return nil, err
		}

		if _, err := s.users.FindUserByID(ctx, id); errors.Is(err, constants.ErrNotFound) {
			out.Skipped++
			continue
		} else if err != nil {
			return nil, err
		}

		if err := s.members.InsertGroupMember(ctx, models.GroupMember{
			GroupID: group.ID, UserID: id, Role: constants.GroupRoleMember,
			JoinMethod: constants.JoinMethodInvite, InvitedByUserID: &inviterID, CreatedAt: now,
		}); err != nil {
			if errors.Is(err, constants.ErrConflict) {
				out.Skipped++
				continue
			}
			return nil, err
		}
		out.Added++
	}

	seenEmails := make(map[string]struct{}, len(in.Emails))
	for _, raw := range in.Emails {
		email, err := validator.NormalizeEmail(raw)
		if err != nil {
			out.Skipped++
			continue
		}
		if _, dup := seenEmails[email]; dup {
			out.Skipped++
			continue
		}
		seenEmails[email] = struct{}{}

		user, err := s.users.FindUserByEmail(ctx, email)
		if err == nil {
			if _, err := s.members.FindGroupMember(ctx, group.ID, user.ID); err == nil {
				out.Skipped++
				continue
			} else if !errors.Is(err, constants.ErrNotFound) {
				return nil, err
			}
			if err := s.members.InsertGroupMember(ctx, models.GroupMember{
				GroupID: group.ID, UserID: user.ID, Role: constants.GroupRoleMember,
				JoinMethod: constants.JoinMethodInvite, InvitedByUserID: &inviterID, CreatedAt: now,
			}); err != nil {
				if errors.Is(err, constants.ErrConflict) {
					out.Skipped++
					continue
				}
				return nil, err
			}
			out.Added++
			continue
		}
		if !errors.Is(err, constants.ErrNotFound) {
			return nil, err
		}

		tokenStr, err := token.NewInviteToken()
		if err != nil {
			return nil, err
		}
		id, err := token.NewID()
		if err != nil {
			return nil, err
		}
		emailCopy := email
		row := models.GroupInviteRow{
			ID:              id,
			Token:           tokenStr,
			GroupID:         group.ID,
			ChannelID:       channelID,
			Email:           &emailCopy,
			InvitedByUserID: inviterID,
			Kind:            constants.GroupInviteKindEmail,
			ExpiresAt:       now.Add(constants.GroupInviteTTL),
			CreatedAt:       now,
		}
		if err := s.invites.InsertInvite(ctx, row); err != nil {
			if errors.Is(err, constants.ErrConflict) {
				out.Skipped++
				continue
			}
			return nil, err
		}

		if err := s.sendInviteEmail(ctx, inviter.DisplayName, group.Name, email, tokenStr); err != nil {
			s.log.Error("invite email failed", "err", err, "to", email)
		}
		out.Emailed++
	}

	if out.Added > 0 || out.Emailed > 0 {
		s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
			GroupID:    group.ID,
			ActorID:    inviterID,
			Category:   constants.GroupAuditCategoryInvite,
			Action:     constants.GroupAuditInviteSent,
			TargetType: constants.GroupAuditTargetInvite,
			Summary:    "Sent group invites",
			Metadata:   map[string]any{"added": out.Added, "emailed": out.Emailed},
			CreatedAt:  now,
		})
	}

	return out, nil
}

func (s service) PreviewInvite(ctx context.Context, inviteToken string) (*models.InvitePreviewResult, error) {
	inviteToken = strings.TrimSpace(inviteToken)
	if inviteToken == "" {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "invite"})
	}

	now := s.now()
	row, err := s.invites.FindByToken(ctx, inviteToken)
	if errors.Is(err, constants.ErrNotFound) {
		return &models.InvitePreviewResult{Valid: false, Reason: constants.InvitePreviewNotFound}, nil
	}
	if err != nil {
		return nil, err
	}

	if row.RedeemedAt != nil {
		return &models.InvitePreviewResult{
			Valid:     false,
			Reason:    constants.InvitePreviewRedeemed,
			ExpiresAt: row.ExpiresAt,
		}, nil
	}
	if !row.ExpiresAt.After(now) {
		return &models.InvitePreviewResult{
			Valid:     false,
			Reason:    constants.InvitePreviewExpired,
			ExpiresAt: row.ExpiresAt,
		}, nil
	}

	group, err := s.groups.FindGroupByID(ctx, row.GroupID)
	if err != nil {
		return nil, err
	}

	channelName := ""
	if !group.IsPrivate && row.ChannelID != nil && strings.TrimSpace(*row.ChannelID) != "" {
		ch, err := s.channels.FindByID(ctx, *row.ChannelID)
		if err == nil {
			channelName = ch.Name
		}
	}

	memberCount, err := s.members.CountGroupMembers(ctx, group.ID)
	if err != nil {
		return nil, err
	}

	onlineCount := 0
	if s.presence != nil && memberCount > 0 {
		if memberIDs, err := s.members.ListGroupMemberIDs(ctx, group.ID); err == nil && len(memberIDs) > 0 {
			if snapshots, err := s.presence.BatchGetEffectivePresence(ctx, memberIDs); err == nil {
				for _, snap := range snapshots {
					if snap.IsOnline {
						onlineCount++
					}
				}
			}
		}
	}

	groupDescription := ""
	groupLogoURL := ""
	groupBannerURL := ""
	groupBannerColor := ""
	if !group.IsPrivate {
		groupDescription = group.Description
		groupLogoURL = objecturl.Public(group.LogoURL, s.objectPublicBase)
		groupBannerURL = objecturl.Public(group.BannerURL, s.objectPublicBase)
		groupBannerColor = group.BannerColor
	}

	return &models.InvitePreviewResult{
		Valid:            true,
		IsPrivate:        group.IsPrivate,
		GroupName:        group.Name,
		GroupDescription: groupDescription,
		ChannelName:      channelName,
		GroupLogoURL:     groupLogoURL,
		GroupBannerURL:   groupBannerURL,
		GroupBannerColor: groupBannerColor,
		MemberCount:      memberCount,
		OnlineCount:      onlineCount,
		ExpiresAt:        row.ExpiresAt,
	}, nil
}

func (s service) RedeemInvite(ctx context.Context, accessToken, inviteToken string) (*models.RedeemInviteResult, error) {
	inviteToken = strings.TrimSpace(inviteToken)
	if inviteToken == "" {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "invite"})
	}

	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, err
	}

	row, err := s.invites.FindByToken(ctx, inviteToken)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "invite"})
		}
		return nil, err
	}
	if row.RedeemedAt != nil || !row.ExpiresAt.After(now) {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "invite"})
	}

	if row.Email != nil && strings.TrimSpace(*row.Email) != "" {
		user, err := s.users.FindUserByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if user.Email == "" || !strings.EqualFold(user.Email, *row.Email) {
			return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
		}
	}

	group, err := s.groups.FindGroupByID(ctx, row.GroupID)
	if err != nil {
		return nil, err
	}

	if banned, err := s.members.IsMemberBanned(ctx, group.ID, userID); err != nil {
		return nil, err
	} else if banned {
		return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	joinedNew := false
	if _, err := s.members.FindGroupMember(ctx, group.ID, userID); err != nil {
		if !errors.Is(err, constants.ErrNotFound) {
			return nil, err
		}
		joinMethod := constants.JoinMethodInvite
		inviteCode := ""
		if row.Kind == constants.GroupInviteKindShareLink {
			joinMethod = constants.JoinMethodInviteLink
			inviteCode = row.Token
		}
		var invitedBy *string
		if row.InvitedByUserID != "" {
			invitedBy = &row.InvitedByUserID
		}
		if err := s.members.InsertGroupMember(ctx, models.GroupMember{
			GroupID: group.ID, UserID: userID, Role: constants.GroupRoleMember,
			JoinMethod: joinMethod, InviteCode: inviteCode, InvitedByUserID: invitedBy, CreatedAt: now,
		}); err != nil && !errors.Is(err, constants.ErrConflict) {
			return nil, err
		}
		joinedNew = true
	}

	if row.Kind == constants.GroupInviteKindEmail {
		if err := s.invites.MarkRedeemed(ctx, row.ID, userID, now); err != nil && !errors.Is(err, constants.ErrNotFound) {
			return nil, err
		}
	}

	if joinedNew {
		s.recordActivity(ctx, models.UserActivityLogEntry{
			UserID:     userID,
			ActorID:    userID,
			Category:   constants.ActivityCategoryGroup,
			Action:     constants.ActivityGroupJoined,
			TargetType: constants.ActivityTargetGroup,
			TargetID:   group.PublicID,
			Summary:    "Joined group via invite",
			Metadata: map[string]any{
				"group_id":    group.ID,
				"group_name":  group.Name,
				"via":         "invite",
				"invite_kind": row.Kind,
			},
			CreatedAt: now,
		})
		s.recordGroupActivity(ctx, models.GroupActivityLogEntry{
			GroupID:    group.ID,
			ActorID:    userID,
			Category:   constants.GroupAuditCategoryMember,
			Action:     constants.GroupAuditMemberJoined,
			TargetType: constants.GroupAuditTargetMember,
			TargetID:   userID,
			Summary:    "Joined group via invite",
			Metadata: map[string]any{
				"via":         "invite",
				"invite_kind": row.Kind,
			},
			CreatedAt: now,
		})
	}

	channelPublic := ""
	if row.ChannelID != nil && strings.TrimSpace(*row.ChannelID) != "" {
		ch, err := s.channels.FindByID(ctx, *row.ChannelID)
		if err == nil {
			channelPublic = ch.PublicID
		}
	}

	return &models.RedeemInviteResult{
		GroupPublicID:   group.PublicID,
		ChannelPublicID: channelPublic,
	}, nil
}

func (s service) sendInviteEmail(ctx context.Context, inviterName, groupName, to, inviteToken string) error {
	if s.notify == nil {
		return nil
	}
	url := s.inviteURL(inviteToken)
	_, err := s.notify.Send(ctx, &notificationv1.SendRequest{
		Channel:    "email",
		TemplateId: "group.invite",
		Locale:     "en",
		Vars: map[string]string{
			"inviter_name": inviterName,
			"group_name":   groupName,
			"invite_url":   url,
		},
		Email: &notificationv1.EmailRecipient{To: to},
	})

	return err
}

func (s service) inviteURL(token string) string {
	if s.webBase == "" {
		return fmt.Sprintf("/invite/%s", token)
	}
	return fmt.Sprintf("%s/invite/%s", s.webBase, token)
}

func (s service) loadGroupMember(ctx context.Context, accessToken, groupPublicID string) (models.Group, models.GroupMember, error) {
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return models.Group{}, models.GroupMember{}, err
	}

	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return models.Group{}, models.GroupMember{}, err
	}

	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return models.Group{}, models.GroupMember{}, exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}
		return models.Group{}, models.GroupMember{}, err
	}

	member, err := s.members.FindGroupMember(ctx, group.ID, userID)
	if errors.Is(err, constants.ErrNotFound) {
		return models.Group{}, models.GroupMember{}, exception.PermissionDenied(exception.MsgForbidden, nil)
	}
	if err != nil {
		return models.Group{}, models.GroupMember{}, err
	}

	return group, member, nil
}

func (s service) resolveChannelID(ctx context.Context, groupID, channelPublicID string) (string, error) {
	channelPublicID = strings.TrimSpace(channelPublicID)
	if channelPublicID == "" {
		return "", nil
	}
	publicID, err := validator.ChannelPublicID(channelPublicID)
	if err != nil {
		return "", err
	}
	ch, err := s.channels.FindByPublicIDInGroup(ctx, groupID, publicID)
	if errors.Is(err, constants.ErrNotFound) {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "channel_id"})
	}
	if err != nil {
		return "", err
	}

	return ch.ID, nil
}

// ValidEmailForChip reports whether free text can become an email invite chip.
func ValidEmailForChip(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	_, err := mail.ParseAddress(raw)
	return err == nil
}
