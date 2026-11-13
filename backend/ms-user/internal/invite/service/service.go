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
	groupmemberrepo "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository"
	grouprolerepo "github.com/wichat/wichat/backend/ms-user/internal/group-role/repository"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	inviterepo "github.com/wichat/wichat/backend/ms-user/internal/invite/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
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
	row, err := s.invites.FindOpenShareLink(ctx, group.ID, channelID, now)
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
			GroupID: group.ID, UserID: id, Role: constants.GroupRoleMember, CreatedAt: now,
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
				GroupID: group.ID, UserID: user.ID, Role: constants.GroupRoleMember, CreatedAt: now,
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
	if row.ChannelID != nil && strings.TrimSpace(*row.ChannelID) != "" {
		ch, err := s.channels.FindByID(ctx, *row.ChannelID)
		if err == nil {
			channelName = ch.Name
		}
	}

	memberCount, err := s.members.CountGroupMembers(ctx, group.ID)
	if err != nil {
		return nil, err
	}

	return &models.InvitePreviewResult{
		Valid:        true,
		GroupName:    group.Name,
		ChannelName:  channelName,
		GroupLogoURL: objecturl.Public(group.LogoURL, s.objectPublicBase),
		MemberCount:  memberCount,
		OnlineCount:  0,
		ExpiresAt:    row.ExpiresAt,
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
			return nil, exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "invite"})
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
		return nil, exception.PermissionDenied("user is banned from this group", nil)
	}

	if _, err := s.members.FindGroupMember(ctx, group.ID, userID); err != nil {
		if !errors.Is(err, constants.ErrNotFound) {
			return nil, err
		}
		if err := s.members.InsertGroupMember(ctx, models.GroupMember{
			GroupID: group.ID, UserID: userID, Role: constants.GroupRoleMember, CreatedAt: now,
		}); err != nil && !errors.Is(err, constants.ErrConflict) {
			return nil, err
		}
	}

	if row.Kind == constants.GroupInviteKindEmail {
		if err := s.invites.MarkRedeemed(ctx, row.ID, userID, now); err != nil && !errors.Is(err, constants.ErrNotFound) {
			return nil, err
		}
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
