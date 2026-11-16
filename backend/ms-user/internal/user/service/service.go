package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	groupmemberrepo "github.com/wichat/wichat/backend/ms-user/internal/group-member/repository"
	grouprepo "github.com/wichat/wichat/backend/ms-user/internal/group/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	activitysvc "github.com/wichat/wichat/backend/ms-user/internal/user-activity-log/service"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	"github.com/wichat/wichat/backend/ms-user/pkg/access"
	authRedis "github.com/wichat/wichat/backend/ms-user/pkg/auth-redis"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/mfa"
	"github.com/wichat/wichat/backend/ms-user/pkg/otp"
	"github.com/wichat/wichat/backend/ms-user/pkg/password"
	"github.com/wichat/wichat/backend/ms-user/pkg/token"
	"github.com/wichat/wichat/backend/ms-user/pkg/username"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	"github.com/wichat/wichat/backend/wi-shared/infra/minio"
	notificationv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/notification/v1"
)

// Notifier sends outbound notification RPCs (email).
type Notifier interface {
	Send(ctx context.Context, in *notificationv1.SendRequest) (*notificationv1.SendResponse, error)
}

// Service is the user/workspace use-case API.
type Service interface {
	SetupStatus(ctx context.Context) (*models.SetupStatus, error)
	CompleteInstanceSetup(ctx context.Context, accessToken string, in models.CompleteInstanceSetupInput) error
	CheckUsername(ctx context.Context, raw string) (models.CheckUsernameResult, error)
	SearchUsers(ctx context.Context, accessToken, query string, limit int, groupPublicID string) ([]models.UserSearchHit, error)
	UploadUserAvatar(ctx context.Context, accessToken, contentType string, body io.Reader, size int64) (*models.PublicUser, error)
	UpdateUserProfile(ctx context.Context, accessToken, displayName, pronouns, bio, bannerColor string, clearAvatar bool) (*models.PublicUser, error)
	UpdateUserStatusNote(ctx context.Context, accessToken, text, emoji, expiresAt string) (*models.PublicUser, error)
	UpdateUserPreferences(ctx context.Context, accessToken, locale, timeFormat string) (*models.PublicUser, error)
	UpdateUsername(ctx context.Context, accessToken, newUsername, currentPassword string) (*models.PublicUser, error)
	RequestEmailChange(ctx context.Context, accessToken, locale string) (*models.AuthChallenge, error)
	VerifyEmailChange(ctx context.Context, accessToken, challengeID, code string) (string, error)
	CompleteEmailChange(ctx context.Context, accessToken, changeToken, newEmail, currentPassword, reason string) (*models.PublicUser, error)
	UpdatePassword(ctx context.Context, accessToken, currentPassword, newPassword string) error
	GetMFAStatus(ctx context.Context, accessToken string) (*models.MFAStatus, error)
	BeginTOTPSetup(ctx context.Context, accessToken, currentPassword string) (*models.TOTPSetupStart, error)
	ConfirmTOTPSetup(ctx context.Context, accessToken, setupToken, code string) ([]string, error)
	DisableTOTP(ctx context.Context, accessToken, currentPassword string) error
	BeginWebAuthnRegister(ctx context.Context, accessToken, currentPassword, name string) (*models.WebAuthnRegisterBegin, error)
	FinishWebAuthnRegister(ctx context.Context, accessToken, challengeID, credentialJSON string) (*models.WebAuthnCredential, []string, error)
	RenameWebAuthnCredential(ctx context.Context, accessToken, credentialID, name string) (*models.WebAuthnCredential, error)
	DeleteWebAuthnCredential(ctx context.Context, accessToken, credentialID, currentPassword string) error
	RegenerateRecoveryCodes(ctx context.Context, accessToken, currentPassword string) ([]string, error)
	DisableAccount(ctx context.Context, accessToken, currentPassword string) error
	DeleteAccount(ctx context.Context, accessToken, currentPassword string) error
	ListSessions(ctx context.Context, accessToken string) ([]models.Session, string, error)
	RevokeSession(ctx context.Context, accessToken, sessionID, currentPassword string) error
	RevokeOtherSessions(ctx context.Context, accessToken, currentPassword string) error
	ListRelationships(ctx context.Context, accessToken string) ([]models.UserRelationship, error)
	SetRelationship(ctx context.Context, accessToken, targetUserID, kind string) (string, error)
	WithActivityLog(r activitysvc.Recorder) Service
}

type service struct {
	users       userrepo.PgRepository
	groups      grouprepo.PgRepository
	members     groupmemberrepo.PgRepository
	objects     userrepo.MinioRepository
	sessions    authrepo.SessionRepository
	refresh     authrepo.RefreshStore
	emailAuth   authrepo.EmailAuthStore
	emailChange userrepo.EmailChangeStore
	mfaStore    userrepo.MFAEnrollmentStore
	webauthn    *webauthn.WebAuthn
	mfaKey      [32]byte
	notify      Notifier
	activity    activitysvc.Recorder
	secret      []byte
	log         logger.Logger
	now         func() time.Time
}

// NewService builds user/workspace use cases.
func NewService(
	users userrepo.PgRepository,
	groups grouprepo.PgRepository,
	members groupmemberrepo.PgRepository,
	objects userrepo.MinioRepository,
	sessions authrepo.SessionRepository,
	emailAuth authrepo.EmailAuthStore,
	emailChange userrepo.EmailChangeStore,
	notify Notifier,
	secret []byte,
	log logger.Logger,
) Service {
	return NewServiceWithMFA(users, groups, members, objects, sessions, nil, emailAuth, emailChange, nil, nil, [32]byte{}, notify, secret, log)
}

// NewServiceWithMFA builds user/workspace use cases with TOTP and WebAuthn MFA support.
func NewServiceWithMFA(
	users userrepo.PgRepository,
	groups grouprepo.PgRepository,
	members groupmemberrepo.PgRepository,
	objects userrepo.MinioRepository,
	sessions authrepo.SessionRepository,
	refresh authrepo.RefreshStore,
	emailAuth authrepo.EmailAuthStore,
	emailChange userrepo.EmailChangeStore,
	mfaStore userrepo.MFAEnrollmentStore,
	wa *webauthn.WebAuthn,
	mfaKey [32]byte,
	notify Notifier,
	secret []byte,
	log logger.Logger,
) Service {
	copied := make([]byte, len(secret))
	copy(copied, secret)

	return service{
		users:       users,
		groups:      groups,
		members:     members,
		objects:     objects,
		sessions:    sessions,
		refresh:     refresh,
		emailAuth:   emailAuth,
		emailChange: emailChange,
		mfaStore:    mfaStore,
		webauthn:    wa,
		mfaKey:      mfaKey,
		notify:      notify,
		secret:      copied,
		log:         log,
		now:         func() time.Time { return time.Now().UTC() },
	}
}

func (s service) WithActivityLog(r activitysvc.Recorder) Service {
	s.activity = r
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

func (s service) SetupStatus(ctx context.Context) (*models.SetupStatus, error) {
	settings, err := s.users.GetInstanceSettings(ctx)
	if err != nil {
		return nil, err
	}

	return &models.SetupStatus{
		RegistrationMode: settings.RegistrationMode,
	}, nil
}

func (s service) CompleteInstanceSetup(ctx context.Context, accessToken string, in models.CompleteInstanceSetupInput) error {
	user, err := access.UserFromAccessToken(ctx, accessToken, s.secret, s.sessions, s.users, s.now())
	if err != nil {
		return err
	}
	if !user.IsInstanceOwner {
		return exception.PermissionDenied(exception.MsgForbidden, nil)
	}

	settings, err := s.users.GetInstanceSettings(ctx)
	if err != nil {
		return err
	}
	if settings.SetupCompleted {
		return exception.Conflict(exception.MsgSetupClosed, nil)
	}

	payload, err := validator.InstanceSetup(in)
	if err != nil {
		return err
	}

	groupID, err := token.NewID()
	if err != nil {
		return exception.Internal(err)
	}
	publicID, err := token.NewPublicID()
	if err != nil {
		return exception.Internal(err)
	}
	payload.GroupID = groupID
	payload.GroupPublicID = publicID

	now := s.now()
	if err := s.users.CompleteInstanceSetup(ctx, user.ID, payload, now); err != nil {
		if errors.Is(err, constants.ErrNotFound) || errors.Is(err, constants.ErrConflict) {
			return exception.Conflict(exception.MsgSetupClosed, nil)
		}

		return err
	}

	return nil
}

func (s service) CheckUsername(ctx context.Context, raw string) (models.CheckUsernameResult, error) {
	normalized, reason, err := username.Validate(raw)
	if err != nil {
		return models.CheckUsernameResult{
			Available:  false,
			Normalized: normalized,
			Reason:     reason,
		}, nil
	}

	_, err = s.users.FindUserByUsername(ctx, normalized)
	if errors.Is(err, constants.ErrNotFound) {
		return models.CheckUsernameResult{
			Available:  true,
			Normalized: normalized,
		}, nil
	}
	if err != nil {
		return models.CheckUsernameResult{}, err
	}

	return models.CheckUsernameResult{
		Available:  false,
		Normalized: normalized,
		Reason:     username.ReasonTaken,
	}, nil
}

func (s service) SearchUsers(
	ctx context.Context,
	accessToken, query string,
	limit int,
	groupPublicID string,
) ([]models.UserSearchHit, error) {
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, err
	}

	publicID, err := validator.GroupPublicID(groupPublicID)
	if err != nil {
		return nil, err
	}

	group, err := s.groups.FindGroupByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "group_id"})
		}
		return nil, err
	}

	if _, err := s.members.FindGroupMember(ctx, group.ID, userID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.PermissionDenied(exception.MsgForbidden, nil)
		}
		return nil, err
	}

	return s.users.SearchUsers(ctx, query, limit, group.ID)
}

func (s service) UploadUserAvatar(
	ctx context.Context,
	accessToken,
	contentType string,
	body io.Reader,
	size int64,
) (*models.PublicUser, error) {
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, err
	}

	if err := validator.UserAvatarUpload(contentType, size); err != nil {
		return nil, err
	}
	if s.objects == nil {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"})
	}

	avatarURL, err := s.objects.PutUserAvatar(ctx, userID, contentType, body, size)
	if err != nil {
		if errors.Is(err, minio.ErrDisabled) {
			return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "file"})
		}

		return nil, err
	}

	if err := s.users.UpdateUserAvatarURL(ctx, userID, avatarURL, now); err != nil {
		return nil, err
	}

	saved, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	pub := saved.Public()
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategoryProfile,
		Action:     constants.ActivityUserAvatarUpdated,
		TargetType: constants.ActivityTargetUser,
		TargetID:   userID,
		Summary:    "Updated avatar",
		Changes:    map[string]models.ActivityFieldChange{"avatar_url": {New: saved.AvatarURL}},
		CreatedAt:  now,
	})
	return &pub, nil
}

func (s service) UpdateUserProfile(
	ctx context.Context,
	accessToken,
	displayName,
	pronouns,
	bio,
	bannerColor string,
	clearAvatar bool,
) (*models.PublicUser, error) {
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, err
	}

	dn, p, b, c, err := validator.UserProfileInput(displayName, pronouns, bio, bannerColor)
	if err != nil {
		return nil, err
	}

	if clearAvatar && s.objects != nil {
		if err := s.objects.RemoveUserAvatar(ctx, userID); err != nil && !errors.Is(err, minio.ErrDisabled) {
			s.log.Warn("remove user avatar object failed", "user_id", userID, "error", err)
		}
	}

	patch := models.UserProfilePatch{
		DisplayName: dn,
		Pronouns:    p,
		Bio:         b,
		BannerColor: c,
		ClearAvatar: clearAvatar,
	}
	if err := s.users.UpdateUserProfile(ctx, userID, patch, now); err != nil {
		return nil, err
	}

	saved, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	pub := saved.Public()
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategoryProfile,
		Action:     constants.ActivityUserProfileUpdated,
		TargetType: constants.ActivityTargetUser,
		TargetID:   userID,
		Summary:    "Updated profile",
		Changes: map[string]models.ActivityFieldChange{
			"display_name": {New: saved.DisplayName},
			"pronouns":     {New: saved.Pronouns},
			"bio":          {New: saved.Bio},
			"banner_color": {New: saved.BannerColor},
			"clear_avatar": {New: clearAvatar},
		},
		CreatedAt: now,
	})
	return &pub, nil
}

func (s service) UpdateUserStatusNote(
	ctx context.Context,
	accessToken,
	text,
	emoji,
	expiresAt string,
) (*models.PublicUser, error) {
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, err
	}

	patch, err := validator.UserStatusNoteInput(text, emoji, expiresAt, now)
	if err != nil {
		return nil, err
	}

	if err := s.users.UpdateUserStatusNote(ctx, userID, patch, now); err != nil {
		return nil, err
	}

	saved, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	pub := saved.Public()
	action := constants.ActivityUserStatusNoteUpdated
	summary := "Updated status note"
	if patch.Text == "" && patch.Emoji == "" {
		action = constants.ActivityUserStatusNoteCleared
		summary = "Cleared status note"
	}
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategoryProfile,
		Action:     action,
		TargetType: constants.ActivityTargetUser,
		TargetID:   userID,
		Summary:    summary,
		Changes: map[string]models.ActivityFieldChange{
			"status_note_text":       {New: patch.Text},
			"status_note_emoji":      {New: patch.Emoji},
			"status_note_expires_at": {New: patch.ExpiresAt},
		},
		CreatedAt: now,
	})
	return &pub, nil
}

func (s service) UpdateUserPreferences(
	ctx context.Context,
	accessToken,
	locale,
	timeFormat string,
) (*models.PublicUser, error) {
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, err
	}

	patch, err := validator.UserPreferencesInput(locale, timeFormat)
	if err != nil {
		return nil, err
	}

	if err := s.users.UpdateUserPreferences(ctx, userID, patch, now); err != nil {
		return nil, err
	}

	saved, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	pub := saved.Public()
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategoryPreferences,
		Action:     constants.ActivityUserPreferencesUpdated,
		TargetType: constants.ActivityTargetUser,
		TargetID:   userID,
		Summary:    "Updated preferences",
		Changes: map[string]models.ActivityFieldChange{
			"locale":      {New: saved.Locale},
			"time_format": {New: saved.TimeFormat},
		},
		CreatedAt: now,
	})
	return &pub, nil
}

func (s service) UpdateUsername(
	ctx context.Context,
	accessToken,
	newUsername,
	currentPassword string,
) (*models.PublicUser, error) {
	// Auth
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, err
	}

	// Input
	normalized, _, err := username.Validate(newUsername)
	if err != nil {
		return nil, err
	}
	if !password.ValidLength(currentPassword) {
		return nil, exception.Unauthenticated(exception.MsgInvalidCredentials, nil)
	}

	// Verify current password
	user, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.PasswordHash == "" || !password.Verify(user.PasswordHash, currentPassword) {
		return nil, exception.Unauthenticated(exception.MsgInvalidCredentials, nil)
	}

	if normalized == user.Username {
		pub := user.Public()
		return &pub, nil
	}

	// Update
	if err := s.users.UpdateUsername(ctx, userID, normalized, now); err != nil {
		if errors.Is(err, constants.ErrConflict) {
			return nil, exception.Conflict(exception.MsgUsernameTaken, nil)
		}

		return nil, err
	}

	saved, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	pub := saved.Public()
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategoryAccount,
		Action:     constants.ActivityUserUsernameUpdated,
		TargetType: constants.ActivityTargetUser,
		TargetID:   userID,
		Summary:    "Updated username",
		Changes:    map[string]models.ActivityFieldChange{"username": {Old: user.Username, New: saved.Username}},
		CreatedAt:  now,
	})
	return &pub, nil
}

func (s service) requireEmailChange() error {
	if s.emailAuth == nil || s.emailChange == nil || s.notify == nil {
		return exception.Unavailable(exception.MsgUnavailable, nil)
	}

	return nil
}

func resetRateLimited(retry time.Duration) error {
	secs := authRedis.CooldownSecs(retry)
	params := map[string]any{"retry_after_seconds": secs}
	if secs < 1 {
		params["retry_after_seconds"] = 1
	}

	return exception.RateLimited(exception.MsgResetRateLimited, params)
}

func maskEmailHint(email string) string {
	email = strings.TrimSpace(email)
	at := strings.LastIndex(email, "@")
	if at <= 1 {
		return "***"
	}
	local, domain := email[:at], email[at:]
	if len(local) <= 2 {
		return local[:1] + "***" + domain
	}

	return local[:1] + "***" + local[len(local)-1:] + domain
}

func (s service) RequestEmailChange(ctx context.Context, accessToken, locale string) (*models.AuthChallenge, error) {
	// Auth
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, err
	}
	if err := s.requireEmailChange(); err != nil {
		return nil, err
	}

	// Rate limit & code generation
	user, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := s.emailAuth.EnsureCanSendEmailCode(ctx, user.Email); err != nil {
		if retry, ok := models.ResetRateLimitRetryAfter(err); ok {
			return nil, resetRateLimited(retry)
		}

		return nil, err
	}

	code, err := otp.Generate(constants.OTPLength)
	if err != nil {
		return nil, exception.Internal(err)
	}
	challengeID, err := token.NewID()
	if err != nil {
		return nil, exception.Internal(err)
	}

	rec := models.EmailChangePending{
		UserID:   user.ID,
		Email:    user.Email,
		CodeHMAC: authRedis.OTPCodeHMAC(s.secret, code),
	}
	if err := s.emailChange.PutChallenge(ctx, challengeID, rec, constants.PasswordResetOTPTTL); err != nil {
		return nil, err
	}

	// Send OTP email
	if _, err := s.notify.Send(ctx, &notificationv1.SendRequest{
		Channel:    "email",
		TemplateId: "auth.email_code",
		Locale:     locale,
		Vars: map[string]string{
			"code":         code,
			"display_name": user.DisplayName,
			"purpose":      "email_change",
		},
		Email: &notificationv1.EmailRecipient{To: user.Email},
	}); err != nil {
		_ = s.emailChange.DeleteChallenge(ctx, challengeID)
		return nil, exception.Internal(err)
	}

	wait, err := s.emailAuth.RecordEmailCodeSent(ctx, user.Email)
	if err != nil {
		return nil, err
	}

	return &models.AuthChallenge{
		ChallengeID:           challengeID,
		EmailHint:             maskEmailHint(user.Email),
		ResendCooldownSeconds: authRedis.CooldownInt32(wait),
	}, nil
}

func (s service) VerifyEmailChange(ctx context.Context, accessToken, challengeID, code string) (string, error) {
	// Auth
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return "", err
	}
	if err := s.requireEmailChange(); err != nil {
		return "", err
	}

	// Input
	challengeID = strings.TrimSpace(challengeID)
	code = strings.TrimSpace(code)
	if challengeID == "" || code == "" {
		return "", exception.InvalidInput(exception.MsgResetCodeInvalid, nil)
	}

	// Verify ownership & OTP
	rec, err := s.emailChange.GetChallenge(ctx, challengeID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return "", exception.InvalidInput(exception.MsgResetCodeInvalid, nil)
		}

		return "", err
	}
	if rec.UserID != userID {
		return "", exception.InvalidInput(exception.MsgResetCodeInvalid, nil)
	}

	verified, err := s.emailChange.VerifyChallenge(ctx, challengeID, code, s.secret, constants.PasswordResetMaxAttempts)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			_ = s.emailAuth.RecordWrongEmailCode(ctx, rec.Email)
			return "", exception.InvalidInput(exception.MsgResetCodeInvalid, nil)
		}

		return "", err
	}

	_ = s.emailChange.DeleteChallenge(ctx, challengeID)
	if err := s.emailAuth.ClearEmailCodePenalty(ctx, verified.Email); err != nil {
		return "", err
	}

	changeToken, err := token.NewID()
	if err != nil {
		return "", exception.Internal(err)
	}
	if err := s.emailChange.SaveChangeToken(ctx, changeToken, userID, constants.PasswordResetTokenTTL); err != nil {
		return "", err
	}

	return changeToken, nil
}

func (s service) CompleteEmailChange(
	ctx context.Context,
	accessToken,
	changeToken,
	newEmail,
	currentPassword,
	reason string,
) (*models.PublicUser, error) {
	// Auth
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, err
	}
	if err := s.requireEmailChange(); err != nil {
		return nil, err
	}

	// Input
	changeToken = strings.TrimSpace(changeToken)
	if changeToken == "" {
		return nil, exception.InvalidInput(exception.MsgResetCodeInvalid, nil)
	}
	normalized, err := validator.NormalizeEmail(newEmail)
	if err != nil {
		return nil, err
	}
	if !password.ValidLength(currentPassword) {
		return nil, exception.Unauthenticated(exception.MsgInvalidCredentials, nil)
	}

	// Verify current password before consuming token
	user, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.PasswordHash == "" || !password.Verify(user.PasswordHash, currentPassword) {
		return nil, exception.Unauthenticated(exception.MsgInvalidCredentials, nil)
	}

	// Consume one-time token
	tokenUserID, err := s.emailChange.ConsumeChangeToken(ctx, changeToken)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.InvalidInput(exception.MsgResetCodeInvalid, nil)
		}

		return nil, err
	}
	if tokenUserID != userID {
		return nil, exception.InvalidInput(exception.MsgResetCodeInvalid, nil)
	}

	if strings.EqualFold(normalized, user.Email) {
		pub := user.Public()
		return &pub, nil
	}

	// Update email
	if err := s.users.UpdateEmail(ctx, userID, normalized, now); err != nil {
		if errors.Is(err, constants.ErrConflict) {
			return nil, exception.Conflict(exception.MsgEmailTaken, nil)
		}

		return nil, err
	}

	s.log.Info("user email changed", "user_id", userID, "reason", strings.TrimSpace(reason))

	saved, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	pub := saved.Public()
	meta := map[string]any(nil)
	if trimmedReason := strings.TrimSpace(reason); trimmedReason != "" {
		meta = map[string]any{"reason": trimmedReason}
	}
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategoryAccount,
		Action:     constants.ActivityUserEmailUpdated,
		TargetType: constants.ActivityTargetUser,
		TargetID:   userID,
		Summary:    "Updated email address",
		Changes:    map[string]models.ActivityFieldChange{"email": {Old: user.Email, New: saved.Email}},
		Metadata:   meta,
		CreatedAt:  now,
	})
	return &pub, nil
}

func (s service) UpdatePassword(ctx context.Context, accessToken, currentPassword, newPassword string) error {
	// Auth
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return err
	}

	// Input
	if err := password.Validate(newPassword); err != nil {
		return err
	}

	// Verify current password when one is already set (OAuth-only users may set a password for the first time).
	user, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.PasswordHash != "" {
		if !password.ValidLength(currentPassword) || !password.Verify(user.PasswordHash, currentPassword) {
			return exception.Unauthenticated(exception.MsgInvalidCredentials, nil)
		}
	}

	// Hash & update
	hash, err := password.Hash(newPassword)
	if err != nil {
		return err
	}

	if err := s.users.UpdateUserPasswordHash(ctx, userID, hash, now); err != nil {
		return err
	}
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategorySecurity,
		Action:     constants.ActivityUserPasswordUpdated,
		TargetType: constants.ActivityTargetUser,
		TargetID:   userID,
		Summary:    "Updated password",
		CreatedAt:  now,
	})
	return nil
}

func (s service) verifyCallerPassword(ctx context.Context, accessToken, currentPassword string) (models.User, time.Time, error) {
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return models.User{}, now, err
	}
	if !password.ValidLength(currentPassword) {
		return models.User{}, now, exception.Unauthenticated(exception.MsgInvalidCredentials, nil)
	}
	user, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		return models.User{}, now, err
	}
	if user.PasswordHash == "" || !password.Verify(user.PasswordHash, currentPassword) {
		return models.User{}, now, exception.Unauthenticated(exception.MsgInvalidCredentials, nil)
	}
	return user, now, nil
}

func (s service) ensureRecoveryCodes(ctx context.Context, userID string, now time.Time) ([]string, error) {
	remaining, err := s.users.CountUnusedRecoveryCodes(ctx, userID)
	if err != nil {
		return nil, err
	}
	if remaining > 0 {
		return nil, nil
	}
	return s.rotateRecoveryCodes(ctx, userID, now)
}

func (s service) rotateRecoveryCodes(ctx context.Context, userID string, now time.Time) ([]string, error) {
	codes, err := mfa.GenerateRecoveryCodes(constants.RecoveryCodeCount)
	if err != nil {
		return nil, exception.Internal(err)
	}
	hashes := make([][]byte, len(codes))
	for i, c := range codes {
		hashes[i] = mfa.HashRecoveryCode(s.secret, c)
	}
	if err := s.users.ReplaceRecoveryCodes(ctx, userID, hashes, now); err != nil {
		return nil, err
	}
	return codes, nil
}

func (s service) cleanupRecoveryCodesIfNoMFA(ctx context.Context, userID string, now time.Time) error {
	totpEnabled := true
	if _, err := s.users.GetMFATOTP(ctx, userID); errors.Is(err, constants.ErrNotFound) {
		totpEnabled = false
	} else if err != nil {
		return err
	}
	keys, err := s.users.ListWebAuthnCredentials(ctx, userID)
	if err != nil {
		return err
	}
	if !totpEnabled && len(keys) == 0 {
		return s.users.ReplaceRecoveryCodes(ctx, userID, nil, now)
	}
	return nil
}

func (s service) GetMFAStatus(ctx context.Context, accessToken string) (*models.MFAStatus, error) {
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, err
	}

	totpEnabled := true
	if _, err := s.users.GetMFATOTP(ctx, userID); errors.Is(err, constants.ErrNotFound) {
		totpEnabled = false
	} else if err != nil {
		return nil, err
	}

	keys, err := s.users.ListWebAuthnCredentials(ctx, userID)
	if err != nil {
		return nil, err
	}

	remaining, err := s.users.CountUnusedRecoveryCodes(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &models.MFAStatus{
		TOTPEnabled:            totpEnabled,
		SecurityKeys:           keys,
		RecoveryCodesRemaining: remaining,
	}, nil
}

func (s service) BeginTOTPSetup(ctx context.Context, accessToken, currentPassword string) (*models.TOTPSetupStart, error) {
	if s.mfaStore == nil {
		return nil, exception.Unavailable(exception.MsgUnavailable, nil)
	}
	user, _, err := s.verifyCallerPassword(ctx, accessToken, currentPassword)
	if err != nil {
		return nil, err
	}

	accountName := user.Email
	if accountName == "" {
		accountName = user.Username
	}
	secret, uri, err := mfa.GenerateTOTP("WiChat", accountName)
	if err != nil {
		return nil, exception.Internal(err)
	}
	setupToken, err := token.NewID()
	if err != nil {
		return nil, exception.Internal(err)
	}

	if err := s.mfaStore.PutTOTPSetup(ctx, setupToken, models.TOTPSetupPending{
		UserID: user.ID,
		Secret: secret,
	}, constants.TOTPSetupTTL); err != nil {
		return nil, err
	}

	return &models.TOTPSetupStart{
		SetupToken: setupToken,
		Secret:     secret,
		OtpauthURI: uri,
	}, nil
}

func (s service) ConfirmTOTPSetup(ctx context.Context, accessToken, setupToken, code string) ([]string, error) {
	if s.mfaStore == nil {
		return nil, exception.Unavailable(exception.MsgUnavailable, nil)
	}
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, err
	}
	setupToken = strings.TrimSpace(setupToken)
	if setupToken == "" || len(strings.TrimSpace(code)) != 6 {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, nil)
	}

	pending, err := s.mfaStore.TakeTOTPSetup(ctx, setupToken)
	if errors.Is(err, constants.ErrNotFound) {
		return nil, exception.InvalidInput(exception.MsgResetCodeInvalid, nil)
	}
	if err != nil {
		return nil, err
	}
	if pending.UserID != userID {
		return nil, exception.InvalidInput(exception.MsgResetCodeInvalid, nil)
	}

	step, ok := mfa.VerifyTOTP(pending.Secret, code, now, 0)
	if !ok {
		return nil, exception.InvalidInput(exception.MsgResetCodeInvalid, nil)
	}

	sealed, err := mfa.EncryptSecret(s.mfaKey, pending.Secret)
	if err != nil {
		return nil, exception.Internal(err)
	}
	if err := s.users.UpsertMFATOTP(ctx, models.MFATOTP{
		UserID:          userID,
		SecretEncrypted: sealed,
		LastUsedStep:    step,
		EnabledAt:       now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}); err != nil {
		return nil, err
	}

	codes, err := s.ensureRecoveryCodes(ctx, userID, now)
	if err != nil {
		return nil, err
	}
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategorySecurity,
		Action:     constants.ActivityUserMFATOTPEnabled,
		TargetType: constants.ActivityTargetUser,
		TargetID:   userID,
		Summary:    "Enabled authenticator app MFA",
		Metadata:   map[string]any{"recovery_codes_generated": len(codes)},
		CreatedAt:  now,
	})
	return codes, nil
}

func (s service) DisableTOTP(ctx context.Context, accessToken, currentPassword string) error {
	user, now, err := s.verifyCallerPassword(ctx, accessToken, currentPassword)
	if err != nil {
		return err
	}
	if err := s.users.DeleteMFATOTP(ctx, user.ID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.NotFound(exception.MsgInvalidInput, nil)
		}
		return err
	}
	if err := s.cleanupRecoveryCodesIfNoMFA(ctx, user.ID, now); err != nil {
		return err
	}
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     user.ID,
		ActorID:    user.ID,
		Category:   constants.ActivityCategorySecurity,
		Action:     constants.ActivityUserMFATOTPDisabled,
		TargetType: constants.ActivityTargetUser,
		TargetID:   user.ID,
		Summary:    "Disabled authenticator app MFA",
		CreatedAt:  now,
	})
	return nil
}

func (s service) BeginWebAuthnRegister(ctx context.Context, accessToken, currentPassword, name string) (*models.WebAuthnRegisterBegin, error) {
	if s.mfaStore == nil || s.webauthn == nil {
		return nil, exception.Unavailable(exception.MsgUnavailable, nil)
	}
	user, _, err := s.verifyCallerPassword(ctx, accessToken, currentPassword)
	if err != nil {
		return nil, err
	}

	cleanName := strings.TrimSpace(name)
	if cleanName == "" || utf8.RuneCountInString(cleanName) > constants.MaxSecurityKeyNameRunes {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "name"})
	}

	existing, err := s.users.ListWebAuthnCredentials(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if len(existing) >= constants.MaxSecurityKeysPerUser {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "security_keys"})
	}

	waCreds := toWebAuthnLibCredentials(existing)
	waUser := mfa.WebAuthnUser{
		ID:          []byte(user.ID),
		Name:        user.Username,
		DisplayName: user.DisplayName,
		Credentials: waCreds,
	}
	creation, sessData, err := s.webauthn.BeginRegistration(
		waUser,
		webauthn.WithExclusions(mfa.CredentialExcludeList(waCreds)),
	)
	if err != nil {
		return nil, exception.Internal(err)
	}

	optsBytes, err := json.Marshal(creation)
	if err != nil {
		return nil, exception.Internal(err)
	}
	sessBytes, err := json.Marshal(sessData)
	if err != nil {
		return nil, exception.Internal(err)
	}
	challengeID, err := token.NewID()
	if err != nil {
		return nil, exception.Internal(err)
	}

	if err := s.mfaStore.PutWebAuthnRegister(ctx, challengeID, models.WebAuthnRegisterPending{
		UserID:      user.ID,
		KeyName:     cleanName,
		SessionJSON: string(sessBytes),
	}, constants.WebAuthnChallengeTTL); err != nil {
		return nil, err
	}

	return &models.WebAuthnRegisterBegin{
		ChallengeID: challengeID,
		OptionsJSON: string(optsBytes),
	}, nil
}

func (s service) FinishWebAuthnRegister(ctx context.Context, accessToken, challengeID, credentialJSON string) (*models.WebAuthnCredential, []string, error) {
	if s.mfaStore == nil || s.webauthn == nil {
		return nil, nil, exception.Unavailable(exception.MsgUnavailable, nil)
	}
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, nil, err
	}
	challengeID = strings.TrimSpace(challengeID)
	if challengeID == "" || strings.TrimSpace(credentialJSON) == "" {
		return nil, nil, exception.InvalidInput(exception.MsgInvalidInput, nil)
	}

	pending, err := s.mfaStore.TakeWebAuthnRegister(ctx, challengeID)
	if errors.Is(err, constants.ErrNotFound) {
		return nil, nil, exception.InvalidInput(exception.MsgResetCodeInvalid, nil)
	}
	if err != nil {
		return nil, nil, err
	}
	if pending.UserID != userID {
		return nil, nil, exception.InvalidInput(exception.MsgResetCodeInvalid, nil)
	}

	var sessData webauthn.SessionData
	if err := json.Unmarshal([]byte(pending.SessionJSON), &sessData); err != nil {
		return nil, nil, exception.Internal(err)
	}

	parsed, err := protocol.ParseCredentialCreationResponseBody(strings.NewReader(credentialJSON))
	if err != nil {
		return nil, nil, exception.InvalidInput(exception.MsgInvalidInput, nil)
	}

	user, err := s.users.FindUserByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	existing, err := s.users.ListWebAuthnCredentials(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	waUser := mfa.WebAuthnUser{
		ID:          []byte(user.ID),
		Name:        user.Username,
		DisplayName: user.DisplayName,
		Credentials: toWebAuthnLibCredentials(existing),
	}

	cred, err := s.webauthn.CreateCredential(waUser, sessData, parsed)
	if err != nil {
		return nil, nil, exception.InvalidInput(exception.MsgInvalidInput, nil)
	}

	id, err := token.NewID()
	if err != nil {
		return nil, nil, exception.Internal(err)
	}
	transports := make([]string, len(cred.Transport))
	for i, tr := range cred.Transport {
		transports[i] = string(tr)
	}

	saved := models.WebAuthnCredential{
		ID:              id,
		UserID:          userID,
		Name:            pending.KeyName,
		CredentialID:    cred.ID,
		PublicKey:       cred.PublicKey,
		AttestationType: cred.AttestationType,
		AAGUID:          cred.Authenticator.AAGUID,
		SignCount:       cred.Authenticator.SignCount,
		CloneWarning:    cred.Authenticator.CloneWarning,
		Transports:      transports,
		BackupEligible:  cred.Flags.BackupEligible,
		BackupState:     cred.Flags.BackupState,
		CreatedAt:       now,
	}
	if err := s.users.InsertWebAuthnCredential(ctx, saved); err != nil {
		if errors.Is(err, constants.ErrConflict) {
			return nil, nil, exception.Conflict(exception.MsgInvalidInput, nil)
		}
		return nil, nil, err
	}

	codes, err := s.ensureRecoveryCodes(ctx, userID, now)
	if err != nil {
		return nil, nil, err
	}

	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategorySecurity,
		Action:     constants.ActivityUserMFAWebAuthnRegistered,
		TargetType: constants.ActivityTargetSecurityKey,
		TargetID:   saved.ID,
		Summary:    "Registered security key",
		Metadata:   map[string]any{"name": saved.Name},
		CreatedAt:  now,
	})
	return &saved, codes, nil
}

func (s service) RenameWebAuthnCredential(ctx context.Context, accessToken, credentialID, name string) (*models.WebAuthnCredential, error) {
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, err
	}
	cleanName := strings.TrimSpace(name)
	if strings.TrimSpace(credentialID) == "" || cleanName == "" || utf8.RuneCountInString(cleanName) > constants.MaxSecurityKeyNameRunes {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "name"})
	}
	updated, err := s.users.RenameWebAuthnCredential(ctx, userID, credentialID, cleanName)
	if errors.Is(err, constants.ErrNotFound) {
		return nil, exception.NotFound(exception.MsgInvalidInput, nil)
	}
	if err != nil {
		return nil, err
	}
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategorySecurity,
		Action:     constants.ActivityUserMFAWebAuthnRenamed,
		TargetType: constants.ActivityTargetSecurityKey,
		TargetID:   updated.ID,
		Summary:    "Renamed security key",
		Changes:    map[string]models.ActivityFieldChange{"name": {New: updated.Name}},
		CreatedAt:  now,
	})
	return &updated, nil
}

func (s service) DeleteWebAuthnCredential(ctx context.Context, accessToken, credentialID, currentPassword string) error {
	user, now, err := s.verifyCallerPassword(ctx, accessToken, currentPassword)
	if err != nil {
		return err
	}
	if strings.TrimSpace(credentialID) == "" {
		return exception.InvalidInput(exception.MsgInvalidInput, nil)
	}
	if err := s.users.DeleteWebAuthnCredential(ctx, user.ID, credentialID); err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return exception.NotFound(exception.MsgInvalidInput, nil)
		}
		return err
	}
	if err := s.cleanupRecoveryCodesIfNoMFA(ctx, user.ID, now); err != nil {
		return err
	}
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     user.ID,
		ActorID:    user.ID,
		Category:   constants.ActivityCategorySecurity,
		Action:     constants.ActivityUserMFAWebAuthnDeleted,
		TargetType: constants.ActivityTargetSecurityKey,
		TargetID:   credentialID,
		Summary:    "Removed security key",
		CreatedAt:  now,
	})
	return nil
}

func (s service) RegenerateRecoveryCodes(ctx context.Context, accessToken, currentPassword string) ([]string, error) {
	user, now, err := s.verifyCallerPassword(ctx, accessToken, currentPassword)
	if err != nil {
		return nil, err
	}
	codes, err := s.rotateRecoveryCodes(ctx, user.ID, now)
	if err != nil {
		return nil, err
	}
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     user.ID,
		ActorID:    user.ID,
		Category:   constants.ActivityCategorySecurity,
		Action:     constants.ActivityUserMFARecoveryRegenerated,
		TargetType: constants.ActivityTargetUser,
		TargetID:   user.ID,
		Summary:    "Regenerated MFA recovery codes",
		Metadata:   map[string]any{"count": len(codes)},
		CreatedAt:  now,
	})
	return codes, nil
}

func (s service) DisableAccount(ctx context.Context, accessToken, currentPassword string) error {
	user, now, err := s.prepareAccountTermination(ctx, accessToken, currentPassword)
	if err != nil {
		return err
	}
	if err := s.users.DisableAccount(ctx, user.ID, now); err != nil {
		return err
	}
	if err := s.revokeAllSessions(ctx, user.ID, now); err != nil {
		return err
	}
	s.log.Info("user account disabled", "user_id", user.ID)
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     user.ID,
		ActorID:    user.ID,
		Category:   constants.ActivityCategoryAccount,
		Action:     constants.ActivityUserAccountDisabled,
		TargetType: constants.ActivityTargetUser,
		TargetID:   user.ID,
		Summary:    "Disabled account",
		CreatedAt:  now,
	})
	return nil
}

func (s service) DeleteAccount(ctx context.Context, accessToken, currentPassword string) error {
	user, now, err := s.prepareAccountTermination(ctx, accessToken, currentPassword)
	if err != nil {
		return err
	}
	compactID := strings.ReplaceAll(user.ID, "-", "")
	if len(compactID) > 16 {
		compactID = compactID[:16]
	}
	anonymized := "deleted_" + compactID
	if err := s.users.SoftDeleteAccount(ctx, user.ID, anonymized, now); err != nil {
		return err
	}
	if err := s.revokeAllSessions(ctx, user.ID, now); err != nil {
		return err
	}
	if user.AvatarURL != "" && s.objects != nil {
		if rmErr := s.objects.RemoveUserAvatar(ctx, user.ID); rmErr != nil {
			s.log.Warn("remove user avatar on delete", "user_id", user.ID, "err", rmErr)
		}
	}
	s.log.Info("user account soft-deleted", "user_id", user.ID)
	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     user.ID,
		ActorID:    user.ID,
		Category:   constants.ActivityCategoryAccount,
		Action:     constants.ActivityUserAccountDeleted,
		TargetType: constants.ActivityTargetUser,
		TargetID:   user.ID,
		Summary:    "Deleted account",
		CreatedAt:  now,
	})
	return nil
}

func (s service) ListSessions(ctx context.Context, accessToken string) ([]models.Session, string, error) {
	now := s.now()
	userID, currentSessionID, err := access.SessionFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, "", err
	}
	sessions, err := s.sessions.ListActiveSessionsForUser(ctx, userID, now)
	if err != nil {
		return nil, "", err
	}

	return sessions, currentSessionID, nil
}

func (s service) RevokeSession(ctx context.Context, accessToken, sessionID, currentPassword string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "session_id"})
	}
	userID, currentSessionID, now, err := s.verifySessionCallerPassword(ctx, accessToken, currentPassword)
	if err != nil {
		return err
	}
	if sessionID == currentSessionID {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "session_id"})
	}

	sess, err := s.sessions.FindSession(ctx, sessionID)
	if errors.Is(err, constants.ErrNotFound) || (err == nil && sess.UserID != userID) {
		return exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "session_id"})
	}
	if err != nil {
		return err
	}
	if sess.RevokedAt != nil {
		return nil
	}

	if err := s.sessions.RevokeSession(ctx, sessionID, now); err != nil {
		return err
	}
	if s.refresh != nil {
		_ = s.refresh.RevokeForSession(ctx, sessionID)
	}

	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategorySession,
		Action:     constants.ActivityUserSessionRevoked,
		TargetType: constants.ActivityTargetSession,
		TargetID:   sessionID,
		Summary:    "Revoked session",
		CreatedAt:  now,
	})
	return nil
}

func (s service) RevokeOtherSessions(ctx context.Context, accessToken, currentPassword string) error {
	userID, currentSessionID, now, err := s.verifySessionCallerPassword(ctx, accessToken, currentPassword)
	if err != nil {
		return err
	}

	ids, err := s.sessions.RevokeOtherSessionsForUser(ctx, userID, currentSessionID, now)
	if err != nil {
		return err
	}
	if s.refresh != nil {
		for _, id := range ids {
			_ = s.refresh.RevokeForSession(ctx, id)
		}
	}

	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategorySession,
		Action:     constants.ActivityUserSessionsRevokedOthers,
		TargetType: constants.ActivityTargetUser,
		TargetID:   userID,
		Summary:    "Revoked other sessions",
		Metadata:   map[string]any{"revoked_count": len(ids)},
		CreatedAt:  now,
	})
	return nil
}

func (s service) verifySessionCallerPassword(ctx context.Context, accessToken, currentPassword string) (string, string, time.Time, error) {
	now := s.now()
	userID, currentSessionID, err := access.SessionFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return "", "", time.Time{}, err
	}

	user, err := s.users.FindUserByID(ctx, userID)
	if errors.Is(err, constants.ErrNotFound) {
		return "", "", time.Time{}, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return "", "", time.Time{}, err
	}
	if user.DeletedAt != nil || user.DisabledAt != nil {
		return "", "", time.Time{}, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	if user.PasswordHash != "" {
		if !password.ValidLength(currentPassword) || !password.Verify(user.PasswordHash, currentPassword) {
			return "", "", time.Time{}, exception.Unauthenticated(exception.MsgInvalidCredentials, nil)
		}
	}

	return userID, currentSessionID, now, nil
}

func (s service) prepareAccountTermination(ctx context.Context, accessToken, currentPassword string) (models.User, time.Time, error) {
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return models.User{}, time.Time{}, err
	}

	user, err := s.users.FindUserByID(ctx, userID)
	if errors.Is(err, constants.ErrNotFound) {
		return models.User{}, time.Time{}, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return models.User{}, time.Time{}, err
	}
	if user.DeletedAt != nil {
		return models.User{}, time.Time{}, exception.Unauthenticated(exception.MsgAccountDeleted, nil)
	}

	if s.groups != nil {
		groups, err := s.groups.ListGroupsForUser(ctx, userID)
		if err != nil {
			return models.User{}, time.Time{}, err
		}
		ownedCount := 0
		for _, g := range groups {
			if g.CallerRole == constants.GroupRoleOwner {
				ownedCount++
			}
		}
		if ownedCount > 0 {
			return models.User{}, time.Time{}, exception.StateConflict(exception.MsgUserOwnsGroups, map[string]any{"count": ownedCount})
		}
	}

	if user.PasswordHash != "" {
		if !password.ValidLength(currentPassword) || !password.Verify(user.PasswordHash, currentPassword) {
			return models.User{}, time.Time{}, exception.Unauthenticated(exception.MsgInvalidCredentials, nil)
		}
	}

	return user, now, nil
}

func (s service) revokeAllSessions(ctx context.Context, userID string, now time.Time) error {
	if s.sessions == nil {
		return nil
	}
	ids, err := s.sessions.ListSessionIDsForUser(ctx, userID)
	if err != nil {
		return err
	}
	if err := s.sessions.RevokeAllSessionsForUser(ctx, userID, now); err != nil {
		return err
	}
	if s.refresh != nil {
		for _, id := range ids {
			_ = s.refresh.RevokeForSession(ctx, id)
		}
	}

	return nil
}

func toWebAuthnLibCredentials(creds []models.WebAuthnCredential) []webauthn.Credential {
	out := make([]webauthn.Credential, len(creds))
	for i, c := range creds {
		transports := make([]protocol.AuthenticatorTransport, len(c.Transports))
		for j, t := range c.Transports {
			transports[j] = protocol.AuthenticatorTransport(t)
		}
		out[i] = webauthn.Credential{
			ID:              c.CredentialID,
			PublicKey:       c.PublicKey,
			AttestationType: c.AttestationType,
			Transport:       transports,
			Flags: webauthn.CredentialFlags{
				BackupEligible: c.BackupEligible,
				BackupState:    c.BackupState,
			},
			Authenticator: webauthn.Authenticator{
				AAGUID:       c.AAGUID,
				SignCount:    c.SignCount,
				CloneWarning: c.CloneWarning,
			},
		}
	}
	return out
}

func (s service) ListRelationships(ctx context.Context, accessToken string) ([]models.UserRelationship, error) {
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return nil, err
	}

	return s.users.ListRelationships(ctx, userID)
}

func (s service) SetRelationship(ctx context.Context, accessToken, targetUserID, kind string) (string, error) {
	now := s.now()
	userID, err := access.UserIDFromAccessToken(ctx, accessToken, s.secret, s.sessions, now)
	if err != nil {
		return "", err
	}

	targetUserID = strings.TrimSpace(targetUserID)
	if targetUserID == "" || targetUserID == userID {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "target_user_id"})
	}

	kind = strings.TrimSpace(strings.ToLower(kind))
	switch kind {
	case "none":
		if err := s.users.DeleteRelationship(ctx, userID, targetUserID); err != nil {
			return "", err
		}
		return "none", nil
	case "ignore", "block":
		if _, err := s.users.FindUserByID(ctx, targetUserID); err != nil {
			if errors.Is(err, constants.ErrNotFound) {
				return "", exception.NotFound(exception.MsgInvalidInput, map[string]any{"field": "target_user_id"})
			}
			return "", err
		}
		if err := s.users.UpsertRelationship(ctx, userID, targetUserID, kind, now); err != nil {
			return "", err
		}
		return kind, nil
	default:
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "kind"})
	}
}
