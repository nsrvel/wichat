package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	inviterepo "github.com/wichat/wichat/backend/ms-user/internal/invite/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	activitysvc "github.com/wichat/wichat/backend/ms-user/internal/user-activity-log/service"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	userservice "github.com/wichat/wichat/backend/ms-user/internal/user/service"
	authRedis "github.com/wichat/wichat/backend/ms-user/pkg/auth-redis"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/jwt"
	"github.com/wichat/wichat/backend/ms-user/pkg/mfa"
	google "github.com/wichat/wichat/backend/ms-user/pkg/oauth"
	"github.com/wichat/wichat/backend/ms-user/pkg/otp"
	"github.com/wichat/wichat/backend/ms-user/pkg/password"
	"github.com/wichat/wichat/backend/ms-user/pkg/token"
	"github.com/wichat/wichat/backend/ms-user/pkg/username"
	"github.com/wichat/wichat/backend/ms-user/pkg/validator"
	"github.com/wichat/wichat/backend/wi-shared/exception"
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	notificationv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/notification/v1"
)

// Notifier sends outbound notification RPCs (email).
type Notifier interface {
	Send(ctx context.Context, in *notificationv1.SendRequest) (*notificationv1.SendResponse, error)
}

// Service is the auth use-case API (register, password login, sessions, Google OAuth).
type Service interface {
	Setup(ctx context.Context, in models.SetupInput) (*models.SessionResult, error)
	Register(ctx context.Context, in models.RegisterInput) (*models.PasswordAuthResult, error)
	Login(ctx context.Context, in models.LoginInput) (*models.PasswordAuthResult, error)
	VerifyRegister(ctx context.Context, challengeID, code string) (*models.SessionResult, error)
	VerifyLogin(ctx context.Context, challengeID, code string) (*models.SessionResult, error)
	ResendRegisterCode(ctx context.Context, challengeID string) (int32, error)
	ResendLoginCode(ctx context.Context, challengeID string) (int32, error)
	Refresh(ctx context.Context, refreshToken string, meta models.ClientMeta) (*models.SessionResult, error)
	Logout(ctx context.Context, refreshToken string) error
	Me(ctx context.Context, accessToken string) (*models.PublicUser, error)
	GoogleStart(ctx context.Context) (string, string, string, error)
	GoogleFinish(ctx context.Context, code, verifier string, meta models.ClientMeta) (*models.SessionResult, error)
	RequestPasswordReset(ctx context.Context, email, locale string) (int, error)
	VerifyPasswordReset(ctx context.Context, email, code string) (string, error)
	CompletePasswordReset(ctx context.Context, resetToken, password string) error
	BeginLoginWebAuthn(ctx context.Context, challengeID string) (string, error)
	VerifyLoginMFA(ctx context.Context, challengeID, method, code, assertionJSON string) (*models.SessionResult, error)
	WithActivityLog(r activitysvc.Recorder) Service
}

type service struct {
	users      userrepo.PgRepository
	sessions   authrepo.SessionRepository
	refresh    authrepo.RefreshStore
	workspace  userservice.Service
	reset      authrepo.PasswordResetStore
	emailAuth  authrepo.EmailAuthStore
	mfa        authrepo.MFAStore
	webauthn   *webauthn.WebAuthn
	mfaKey     [32]byte
	notify     Notifier
	activity   activitysvc.Recorder
	devOTP     authrepo.DevOTPStore
	devE2EOTP  bool
	log        logger.Logger
	secret     []byte
	now        func() time.Time
	dummy      string
	google     google.Config
	inviteOnly bool
	invites    inviterepo.PgRepository
}

// Options overrides auth service wiring (tests and optional MFA).
type Options struct {
	Now        func() time.Time
	MFA        authrepo.MFAStore
	WebAuthn   *webauthn.WebAuthn
	MFAKey     [32]byte
	DevE2EOTP  bool
	DevOTP     authrepo.DevOTPStore
	InviteOnly bool
	Invites    inviterepo.PgRepository
}

// NewService builds the auth use cases. secret is the HMAC key and must be at least 32 bytes.
func NewService(
	users userrepo.PgRepository,
	sessions authrepo.SessionRepository,
	refresh authrepo.RefreshStore,
	secret []byte,
	log logger.Logger,
	googleCfg google.Config,
	reset authrepo.PasswordResetStore,
	emailAuth authrepo.EmailAuthStore,
	notify Notifier,
	workspace userservice.Service,
	opts *Options,
) (Service, error) {
	if len(secret) < constants.MinJWTSecretLen {
		return nil, errors.New("auth: jwt secret must be at least 32 bytes")
	}

	dummy, err := password.DummyHash()
	if err != nil {
		return nil, err
	}

	copied := make([]byte, len(secret))
	copy(copied, secret)

	now := func() time.Time { return time.Now().UTC() }
	var mfaStore authrepo.MFAStore
	var wa *webauthn.WebAuthn
	var mfaKey [32]byte
	if opts != nil {
		if opts.Now != nil {
			now = opts.Now
		}
		mfaStore = opts.MFA
		wa = opts.WebAuthn
		mfaKey = opts.MFAKey
	}

	devE2EOTP := false
	var devOTP authrepo.DevOTPStore
	inviteOnly := false
	var invites inviterepo.PgRepository
	if opts != nil {
		devE2EOTP = opts.DevE2EOTP
		devOTP = opts.DevOTP
		inviteOnly = opts.InviteOnly
		invites = opts.Invites
	}

	return service{
		users:      users,
		sessions:   sessions,
		refresh:    refresh,
		workspace:  workspace,
		reset:      reset,
		emailAuth:  emailAuth,
		mfa:        mfaStore,
		webauthn:   wa,
		mfaKey:     mfaKey,
		notify:     notify,
		devOTP:     devOTP,
		devE2EOTP:  devE2EOTP,
		log:        log,
		secret:     copied,
		now:        now,
		dummy:      dummy,
		google:     google.Prepare(googleCfg),
		inviteOnly: inviteOnly,
		invites:    invites,
	}, nil
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

// Setup creates the first user via the register + email OTP flow (legacy RPC).
func (s service) Setup(ctx context.Context, in models.SetupInput) (*models.SessionResult, error) {
	out, err := s.Register(ctx, models.RegisterInput{
		Email:       in.Email,
		Password:    in.Password,
		DisplayName: in.DisplayName,
		ClientMeta:  in.ClientMeta,
	})
	if err != nil {
		return nil, err
	}
	if out.Session == nil {
		return nil, exception.Unavailable(exception.MsgUnavailable, nil)
	}

	return out.Session, nil
}

// Register validates signup and starts the email OTP challenge.
func (s service) Register(ctx context.Context, in models.RegisterInput) (*models.PasswordAuthResult, error) {
	email, handle, name, birth, err := validator.RegisterInput(in)
	if err != nil {
		return nil, err
	}

	if err := s.assertRegistrationAllowed(ctx, email, in.InviteToken); err != nil {
		return nil, err
	}
	if err := s.requireEmailAuth(); err != nil {
		return nil, err
	}

	if err := s.emailAuth.EnsureNoActiveRegisterChallenge(ctx, email); err != nil {
		if errors.Is(err, constants.ErrRegisterChallengeActive) {
			return nil, registerUnavailable()
		}
		return nil, err
	}

	if _, err := s.users.FindUserByEmail(ctx, email); err == nil {
		return nil, registerUnavailable()
	} else if !errors.Is(err, constants.ErrNotFound) {
		return nil, err
	}
	if _, err := s.users.FindUserByUsername(ctx, handle); err == nil {
		return nil, registerUnavailable()
	} else if !errors.Is(err, constants.ErrNotFound) {
		return nil, err
	}

	hash, err := password.Hash(in.Password)
	if err != nil {
		return nil, exception.Internal(err)
	}

	birthStr := in.BirthDate
	if birth != nil {
		birthStr = birth.Format("2006-01-02")
	}

	ch, code, err := s.beginEmailChallenge(ctx, email, name, in.Locale, "sign_up")
	if err != nil {
		return nil, err
	}

	pending := models.RegisterPending{
		Email:        email,
		Username:     handle,
		PasswordHash: hash,
		DisplayName:  name,
		BirthDate:    birthStr,
		Locale:       strings.TrimSpace(in.Locale),
		CodeHMAC:     authRedis.OTPCodeHMAC(s.secret, code),
		UserAgent:    in.UserAgent,
		ClientIP:     in.ClientIP,
	}
	if err := s.emailAuth.PutRegisterPending(ctx, ch.ChallengeID, pending, constants.EmailAuthOTPTTL); err != nil {
		if errors.Is(err, constants.ErrRegisterChallengeActive) {
			return nil, registerUnavailable()
		}
		return nil, err
	}

	if err := s.sendEmailCode(ctx, email, name, in.Locale, "sign_up", code); err != nil {
		_ = s.emailAuth.DeleteRegisterPending(ctx, ch.ChallengeID)
		return nil, err
	}
	s.storeDevE2EOTP(ctx, ch.ChallengeID, code)
	if err := s.markEmailCodeSent(ctx, ch, email); err != nil {
		return nil, err
	}

	return &models.PasswordAuthResult{Challenge: ch}, nil
}

func (s service) createUserFromPending(ctx context.Context, pending models.RegisterPending) (models.User, error) {
	if _, err := s.users.FindUserByEmail(ctx, pending.Email); err == nil {
		return models.User{}, registerUnavailable()
	} else if !errors.Is(err, constants.ErrNotFound) {
		return models.User{}, err
	}

	birth, err := validator.ParseBirthDate(pending.BirthDate)
	if err != nil {
		return models.User{}, err
	}

	n, err := s.users.CountUsers(ctx)
	if err != nil {
		return models.User{}, err
	}
	firstAccount := n == 0

	handle, err := username.Normalize(pending.Username)
	if err != nil {
		return models.User{}, err
	}
	if _, err := s.users.FindUserByUsername(ctx, handle); err == nil {
		return models.User{}, registerUnavailable()
	} else if !errors.Is(err, constants.ErrNotFound) {
		return models.User{}, err
	}

	now := s.now()
	id, err := token.NewID()
	if err != nil {
		return models.User{}, exception.Internal(err)
	}

	user := models.User{
		ID:              id,
		Email:           pending.Email,
		Username:        handle,
		PasswordHash:    pending.PasswordHash,
		DisplayName:     pending.DisplayName,
		BirthDate:       birth,
		IsGuest:         false,
		IsInstanceOwner: firstAccount,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if firstAccount {
		if err := s.users.CreateFirstUser(ctx, user); err != nil {
			if errors.Is(err, constants.ErrConflict) {
				return models.User{}, exception.Conflict(exception.MsgSetupClosed, nil)
			}

			return models.User{}, err
		}
	} else if err := s.users.CreateUser(ctx, user); err != nil {
		if errors.Is(err, constants.ErrConflict) {
			return models.User{}, registerUnavailable()
		}

		return models.User{}, err
	}

	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     user.ID,
		ActorID:    user.ID,
		Category:   constants.ActivityCategoryAuth,
		Action:     constants.ActivityUserRegistered,
		TargetType: constants.ActivityTargetUser,
		TargetID:   user.ID,
		Summary:    "Registered account",
		Metadata:   map[string]any{"is_instance_owner": user.IsInstanceOwner},
		CreatedAt:  now,
	})
	return user, nil
}

func (s service) Login(ctx context.Context, in models.LoginInput) (*models.PasswordAuthResult, error) {
	login := strings.TrimSpace(in.Login)
	if login == "" {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "login"})
	}
	if !password.ValidLength(in.Password) {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "password"})
	}

	if s.emailAuth != nil {
		if err := s.emailAuth.EnsureCanAttemptLogin(ctx, login); err != nil {
			if retry, ok := models.ResetRateLimitRetryAfter(err); ok {
				return nil, resetRateLimited(retry)
			}
			return nil, err
		}
	}

	user, findErr := s.users.FindUserByLogin(ctx, login)
	hash := s.dummy
	if findErr == nil && user.PasswordHash != "" {
		hash = user.PasswordHash
	}
	if !password.Verify(hash, in.Password) || findErr != nil || user.PasswordHash == "" {
		if s.emailAuth != nil {
			_, _ = s.emailAuth.RecordFailedLogin(ctx, login)
		}
		return nil, exception.Unauthenticated(exception.MsgInvalidCredentials, nil)
	}
	if s.emailAuth != nil {
		_ = s.emailAuth.ClearFailedLogin(ctx, login)
	}
	if err := ensureAccountActive(user); err != nil {
		return nil, exception.Unauthenticated(exception.MsgInvalidCredentials, nil)
	}

	if res, ok, err := s.maybeBeginLoginMFA(ctx, user, in.UserAgent, in.ClientIP); err != nil {
		return nil, err
	} else if ok {
		return res, nil
	}

	ch, code, err := s.beginEmailChallenge(ctx, user.Email, user.DisplayName, in.Locale, "sign_in")
	if err != nil {
		return nil, err
	}

	pending := models.LoginPending{
		UserID:      user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Locale:      strings.TrimSpace(in.Locale),
		CodeHMAC:    authRedis.OTPCodeHMAC(s.secret, code),
		UserAgent:   in.UserAgent,
		ClientIP:    in.ClientIP,
	}
	if err := s.emailAuth.PutLoginChallenge(ctx, ch.ChallengeID, pending, constants.EmailAuthOTPTTL); err != nil {
		return nil, err
	}

	if err := s.sendEmailCode(ctx, user.Email, user.DisplayName, in.Locale, "sign_in", code); err != nil {
		_ = s.emailAuth.DeleteLoginChallenge(ctx, ch.ChallengeID)
		return nil, err
	}
	s.storeDevE2EOTP(ctx, ch.ChallengeID, code)
	if err := s.markEmailCodeSent(ctx, ch, user.Email); err != nil {
		return nil, err
	}

	return &models.PasswordAuthResult{Challenge: ch}, nil
}

func (s service) requireEmailAuth() error {
	if s.emailAuth == nil || s.notify == nil {
		return exception.Unavailable(exception.MsgUnavailable, nil)
	}

	return nil
}

func (s service) requirePasswordReset() error {
	if s.reset == nil || s.notify == nil {
		return exception.Unavailable(exception.MsgUnavailable, nil)
	}

	return nil
}

func (s service) sendEmailCode(ctx context.Context, to, displayName, locale, purpose, code string) error {
	_, err := s.notify.Send(ctx, &notificationv1.SendRequest{
		Channel:    "email",
		TemplateId: "auth.email_code",
		Locale:     locale,
		Vars: map[string]string{
			"code":         code,
			"display_name": displayName,
			"purpose":      purpose,
		},
		Email: &notificationv1.EmailRecipient{To: to},
	})

	return err
}

func (s service) sendPasswordResetCode(ctx context.Context, to, displayName, locale, code string) error {
	_, err := s.notify.Send(ctx, &notificationv1.SendRequest{
		Channel:    "email",
		TemplateId: "auth.password_reset",
		Locale:     locale,
		Vars: map[string]string{
			"code":            code,
			"display_name":    displayName,
			"expires_minutes": "10",
		},
		Email: &notificationv1.EmailRecipient{To: to},
	})

	return err
}

func (s service) beginEmailChallenge(ctx context.Context, email, displayName, locale, purpose string) (*models.AuthChallenge, string, error) {
	if err := s.requireEmailAuth(); err != nil {
		return nil, "", err
	}
	email = strings.TrimSpace(email)
	if err := s.emailAuth.EnsureCanSendEmailCode(ctx, email); err != nil {
		if retry, ok := models.ResetRateLimitRetryAfter(err); ok {
			return nil, "", resetRateLimited(retry)
		}

		return nil, "", err
	}

	code, err := otp.Generate(constants.OTPLength)
	if err != nil {
		return nil, "", exception.Internal(err)
	}

	challengeID, err := token.NewID()
	if err != nil {
		return nil, "", exception.Internal(err)
	}

	emailHint := "***"
	if at := strings.LastIndex(email, "@"); at > 1 {
		local, domain := email[:at], email[at:]
		if len(local) <= 2 {
			emailHint = local[:1] + "***" + domain
		} else {
			emailHint = local[:1] + "***" + local[len(local)-1:] + domain
		}
	}

	ch := &models.AuthChallenge{
		ChallengeID: challengeID,
		EmailHint:   emailHint,
	}

	return ch, code, nil
}

func (s service) storeDevE2EOTP(ctx context.Context, challengeID, code string) {
	if !s.devE2EOTP || s.devOTP == nil {
		return
	}
	if err := s.devOTP.Put(ctx, challengeID, code, constants.EmailAuthOTPTTL); err != nil {
		s.log.Warn("dev e2e otp shadow not stored", "err", err, "challenge_id", challengeID)
	}
}

func (s service) markEmailCodeSent(ctx context.Context, ch *models.AuthChallenge, email string) error {
	wait, err := s.emailAuth.RecordEmailCodeSent(ctx, strings.TrimSpace(email))
	if err != nil {
		return err
	}
	ch.ResendCooldownSeconds = authRedis.CooldownInt32(wait)

	return nil
}

func (s service) resendEmailCode(
	ctx context.Context,
	challengeID, email, displayName, locale, purpose string,
	update func(codeHMAC string) error,
) (int32, error) {
	if err := s.requireEmailAuth(); err != nil {
		return 0, err
	}
	email = strings.TrimSpace(email)
	if err := s.emailAuth.EnsureCanSendEmailCode(ctx, email); err != nil {
		if retry, ok := models.ResetRateLimitRetryAfter(err); ok {
			return authRedis.CooldownInt32(retry), resetRateLimited(retry)
		}

		return 0, err
	}

	code, err := otp.Generate(constants.OTPLength)
	if err != nil {
		return 0, exception.Internal(err)
	}

	if err := update(authRedis.OTPCodeHMAC(s.secret, code)); err != nil {
		return 0, err
	}

	if err := s.sendEmailCode(ctx, email, displayName, locale, purpose, code); err != nil {
		return 0, err
	}
	s.storeDevE2EOTP(ctx, challengeID, code)

	wait, err := s.emailAuth.RecordEmailCodeSent(ctx, email)
	if err != nil {
		return 0, err
	}

	return authRedis.CooldownInt32(wait), nil
}

// VerifyRegister completes signup after email OTP.
func (s service) VerifyRegister(ctx context.Context, challengeID, code string) (*models.SessionResult, error) {
	challengeID = strings.TrimSpace(challengeID)
	code = strings.TrimSpace(code)
	if challengeID == "" || code == "" {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, nil)
	}
	if err := s.requireEmailAuth(); err != nil {
		return nil, err
	}

	pending, err := s.emailAuth.VerifyRegisterCode(ctx, challengeID, code, s.secret, constants.EmailAuthMaxAttempts)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			email := ""
			if rec, getErr := s.emailAuth.GetRegisterPending(ctx, challengeID); getErr == nil {
				email = rec.Email
			}
			if email != "" {
				_ = s.emailAuth.RecordWrongEmailCode(ctx, email)
			}
			return nil, exception.Unauthenticated(exception.MsgResetCodeInvalid, nil)
		}

		return nil, err
	}

	user, err := s.createUserFromPending(ctx, pending)
	if err != nil {
		return nil, err
	}

	if err := s.emailAuth.DeleteRegisterPending(ctx, challengeID); err != nil {
		return nil, err
	}
	if err := s.emailAuth.ClearEmailCodePenalty(ctx, pending.Email); err != nil {
		return nil, err
	}

	return s.openSession(ctx, user, models.ClientMeta{UserAgent: pending.UserAgent, ClientIP: pending.ClientIP})
}

// VerifyLogin completes password login after email OTP.
func (s service) VerifyLogin(ctx context.Context, challengeID, code string) (*models.SessionResult, error) {
	challengeID = strings.TrimSpace(challengeID)
	code = strings.TrimSpace(code)
	if challengeID == "" || code == "" {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, nil)
	}
	if err := s.requireEmailAuth(); err != nil {
		return nil, err
	}

	pending, err := s.emailAuth.VerifyLoginCode(ctx, challengeID, code, s.secret, constants.EmailAuthMaxAttempts)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			if rec, getErr := s.emailAuth.GetLoginChallenge(ctx, challengeID); getErr == nil {
				_ = s.emailAuth.RecordWrongEmailCode(ctx, rec.Email)
			}
			return nil, exception.Unauthenticated(exception.MsgResetCodeInvalid, nil)
		}

		return nil, err
	}

	user, err := s.users.FindUserByID(ctx, pending.UserID)
	if err != nil {
		return nil, err
	}

	sess, err := s.openSession(ctx, user, models.ClientMeta{UserAgent: pending.UserAgent, ClientIP: pending.ClientIP})
	if err != nil {
		return nil, err
	}
	if err := s.emailAuth.ClearEmailCodePenalty(ctx, pending.Email); err != nil {
		return nil, err
	}

	return sess, nil
}

// ResendRegisterCode emails a new signup OTP for an existing challenge.
func (s service) ResendRegisterCode(ctx context.Context, challengeID string) (int32, error) {
	challengeID = strings.TrimSpace(challengeID)
	if challengeID == "" {
		return 0, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "challenge_id"})
	}

	pending, err := s.emailAuth.GetRegisterPending(ctx, challengeID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return 0, exception.Unauthenticated(exception.MsgResetCodeInvalid, nil)
		}

		return 0, err
	}

	return s.resendEmailCode(ctx, challengeID, pending.Email, pending.DisplayName, pending.Locale, "sign_up", func(hmac string) error {
		return s.emailAuth.UpdateRegisterCode(ctx, challengeID, hmac, constants.EmailAuthOTPTTL)
	})
}

// ResendLoginCode emails a new login OTP for an existing challenge.
func (s service) ResendLoginCode(ctx context.Context, challengeID string) (int32, error) {
	challengeID = strings.TrimSpace(challengeID)
	if challengeID == "" {
		return 0, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "challenge_id"})
	}

	pending, err := s.emailAuth.GetLoginChallenge(ctx, challengeID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return 0, exception.Unauthenticated(exception.MsgResetCodeInvalid, nil)
		}

		return 0, err
	}

	return s.resendEmailCode(ctx, challengeID, pending.Email, pending.DisplayName, pending.Locale, "sign_in", func(hmac string) error {
		return s.emailAuth.UpdateLoginCode(ctx, challengeID, hmac, constants.EmailAuthOTPTTL)
	})
}

// RequestPasswordReset starts password reset when the account exists (generic success otherwise).
func (s service) RequestPasswordReset(ctx context.Context, emailRaw, locale string) (int, error) {
	email, err := validator.NormalizeEmail(emailRaw)
	if err != nil {
		return 0, err
	}

	user, err := s.users.FindUserByEmail(ctx, email)
	if errors.Is(err, constants.ErrNotFound) {
		return s.acknowledgePasswordResetRequest(ctx, email)
	}
	if err != nil {
		return 0, err
	}

	if err := s.requirePasswordReset(); err != nil {
		return 0, err
	}

	if err := s.reset.EnsureCanSendResetCode(ctx, email); err != nil {
		if retry, ok := models.ResetRateLimitRetryAfter(err); ok {
			return authRedis.CooldownSecs(retry), resetRateLimited(retry)
		}

		return 0, err
	}

	code, err := otp.Generate(constants.OTPLength)
	if err != nil {
		return 0, exception.Internal(err)
	}

	if err := s.reset.SaveOTP(ctx, email, user.ID, code, s.secret, constants.PasswordResetOTPTTL); err != nil {
		return 0, err
	}

	if err := s.sendPasswordResetCode(ctx, email, user.DisplayName, locale, code); err != nil {
		_ = s.reset.DeleteOTP(ctx, email)
		return 0, err
	}
	s.storeDevE2EOTP(ctx, authRedis.EmailKey(email), code)

	wait, err := s.reset.RecordResetCodeSent(ctx, email)
	if err != nil {
		return 0, err
	}

	return authRedis.CooldownSecs(wait), nil
}

// acknowledgePasswordResetRequest applies the same resend limits without emailing (anti-enumeration).
func (s service) acknowledgePasswordResetRequest(ctx context.Context, email string) (int, error) {
	if err := s.requirePasswordReset(); err != nil {
		return 0, err
	}

	if err := s.reset.EnsureCanSendResetCode(ctx, email); err != nil {
		if retry, ok := models.ResetRateLimitRetryAfter(err); ok {
			return authRedis.CooldownSecs(retry), resetRateLimited(retry)
		}

		return 0, err
	}

	wait, err := s.reset.RecordResetCodeSent(ctx, email)
	if err != nil {
		return 0, err
	}

	return authRedis.CooldownSecs(wait), nil
}

// VerifyPasswordReset checks the emailed code and returns a one-time reset token.
func (s service) VerifyPasswordReset(ctx context.Context, emailRaw, code string) (string, error) {
	email, err := validator.NormalizeEmail(emailRaw)
	if err != nil {
		return "", err
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return "", exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "code"})
	}
	if err := s.requirePasswordReset(); err != nil {
		return "", err
	}

	userID, err := s.reset.VerifyOTP(ctx, email, code, s.secret, constants.PasswordResetMaxAttempts)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			_ = s.reset.RecordWrongResetCode(ctx, email)
			return "", exception.Unauthenticated(exception.MsgResetCodeInvalid, nil)
		}

		return "", err
	}

	if err := s.reset.ClearResendPenalty(ctx, email); err != nil {
		return "", err
	}

	resetToken, err := token.NewID()
	if err != nil {
		return "", exception.Internal(err)
	}

	if err := s.reset.SaveResetToken(ctx, resetToken, userID, email, constants.PasswordResetTokenTTL); err != nil {
		return "", err
	}

	return resetToken, nil
}

// CompletePasswordReset sets a new password from a reset token.
func (s service) CompletePasswordReset(ctx context.Context, resetToken, newPassword string) error {
	resetToken = strings.TrimSpace(resetToken)
	if resetToken == "" {
		return exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "reset_token"})
	}
	if err := password.Validate(newPassword); err != nil {
		return err
	}
	if err := s.requirePasswordReset(); err != nil {
		return err
	}

	userID, email, err := s.reset.GetResetToken(ctx, resetToken)
	if errors.Is(err, constants.ErrNotFound) {
		return exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return err
	}

	hash, err := password.Hash(newPassword)
	if err != nil {
		return exception.Internal(err)
	}

	now := s.now()
	if err := s.users.UpdateUserPasswordHash(ctx, userID, hash, now); err != nil {
		return err
	}

	if err := s.revokeAllUserSessions(ctx, userID, now); err != nil {
		return err
	}

	if err := s.reset.DeleteResetToken(ctx, resetToken); err != nil {
		return err
	}

	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     userID,
		ActorID:    userID,
		Category:   constants.ActivityCategoryAuth,
		Action:     constants.ActivityUserPasswordReset,
		TargetType: constants.ActivityTargetUser,
		TargetID:   userID,
		Summary:    "Completed password reset",
		Metadata:   map[string]any{"email": email},
		CreatedAt:  now,
	})

	return nil
}

// Refresh rotates the refresh token and issues a new access JWT for the same session id.
func (s service) Refresh(ctx context.Context, refreshToken string, _ models.ClientMeta) (*models.SessionResult, error) {
	if refreshToken == "" {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	currentHash := token.HashRefresh(refreshToken)
	now := s.now()

	rec, err := s.refresh.FindByHash(ctx, currentHash)
	if err != nil && !errors.Is(err, constants.ErrNotFound) {
		return nil, err
	}

	if errors.Is(err, constants.ErrNotFound) {
		sess, err := s.sessions.FindSessionByRefreshHash(ctx, currentHash, now)
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
		}
		if err != nil {
			return nil, err
		}

		return s.finishRefreshRotation(ctx, sess, currentHash)
	}

	if rec.UsedAt != nil {
		return s.refreshReuseResponse(ctx, rec.SessionID, rec.UsedAt)
	}

	sess, err := s.sessions.FindSession(ctx, rec.SessionID)
	if errors.Is(err, constants.ErrNotFound) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return nil, err
	}

	return s.finishRefreshRotation(ctx, sess, currentHash)
}

func (s service) finishRefreshRotation(ctx context.Context, sess models.Session, currentHash []byte) (*models.SessionResult, error) {
	if sess.RevokedAt != nil || !sess.ExpiresAt.After(s.now()) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	plain, nextHash, err := token.NewRefresh()
	if err != nil {
		return nil, exception.Internal(err)
	}
	nextID, err := token.NewID()
	if err != nil {
		return nil, exception.Internal(err)
	}

	now := s.now()
	sessionExp := now.Add(constants.RefreshTokenTTL)
	nextRefresh := models.RefreshToken{
		ID: nextID, SessionID: sess.ID, TokenHash: nextHash, CreatedAt: now,
	}
	err = s.refresh.Rotate(ctx, currentHash, nextHash, nextRefresh, sess.UserID, sessionExp, now)
	if errors.Is(err, constants.ErrAlreadyUsed) {
		return s.refreshReuseResponse(ctx, sess.ID, &now)
	}
	if errors.Is(err, constants.ErrNotFound) {
		if len(sess.RefreshTokenHash) == 0 || !bytes.Equal(sess.RefreshTokenHash, currentHash) {
			return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
		}
		if err := s.refresh.Put(ctx, nextRefresh, sess.UserID, sessionExp); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	if err := s.sessions.TouchSession(ctx, sess.ID, nextID, nextHash, sessionExp, now); err != nil {
		return nil, err
	}

	user, err := s.users.FindUserByID(ctx, sess.UserID)
	if errors.Is(err, constants.ErrNotFound) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return nil, err
	}
	if err := ensureAccountActive(user); err != nil {
		return nil, err
	}

	access, expiresIn, err := jwt.Issue(s.secret, user.ID, sess.ID, now, constants.AccessTokenTTL)
	if err != nil {
		return nil, exception.Internal(err)
	}

	return &models.SessionResult{
		AccessToken:  access,
		ExpiresIn:    expiresIn,
		RefreshToken: plain,
		User:         user.Public(),
	}, nil
}

// Logout revokes the session tied to the refresh token, if any.
func (s service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	hash := token.HashRefresh(refreshToken)
	rec, err := s.refresh.FindByHash(ctx, hash)
	if errors.Is(err, constants.ErrNotFound) {
		sess, err := s.sessions.FindSessionByRefreshHash(ctx, hash, s.now())
		if errors.Is(err, constants.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		_ = s.refresh.RevokeForSession(ctx, sess.ID)

		if err := s.sessions.RevokeSession(ctx, sess.ID, s.now()); err != nil {
			return err
		}
		s.recordActivity(ctx, models.UserActivityLogEntry{
			UserID:     sess.UserID,
			ActorID:    sess.UserID,
			Category:   constants.ActivityCategoryAuth,
			Action:     constants.ActivityUserLoggedOut,
			TargetType: constants.ActivityTargetSession,
			TargetID:   sess.ID,
			Summary:    "Logged out",
		})
		return nil
	}
	if err != nil {
		return err
	}

	var loggedOutUserID string
	if sess, findErr := s.sessions.FindSession(ctx, rec.SessionID); findErr == nil {
		loggedOutUserID = sess.UserID
	}
	_ = s.refresh.RevokeForSession(ctx, rec.SessionID)

	if err := s.sessions.RevokeSession(ctx, rec.SessionID, s.now()); err != nil {
		return err
	}
	if loggedOutUserID != "" {
		s.recordActivity(ctx, models.UserActivityLogEntry{
			UserID:     loggedOutUserID,
			ActorID:    loggedOutUserID,
			Category:   constants.ActivityCategoryAuth,
			Action:     constants.ActivityUserLoggedOut,
			TargetType: constants.ActivityTargetSession,
			TargetID:   rec.SessionID,
			Summary:    "Logged out",
		})
	}
	return nil
}

// Me loads the user for a valid access token and live session.
func (s service) Me(ctx context.Context, accessToken string) (*models.PublicUser, error) {
	if accessToken == "" {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	userID, sessionID, err := jwt.Parse(s.secret, accessToken)
	if err != nil {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	sess, err := s.sessions.FindSession(ctx, sessionID)
	if errors.Is(err, constants.ErrNotFound) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return nil, err
	}

	if sess.UserID != userID || sess.RevokedAt != nil || !sess.ExpiresAt.After(s.now()) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	user, err := s.users.FindUserByID(ctx, userID)
	if errors.Is(err, constants.ErrNotFound) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return nil, err
	}

	pub := user.Public()

	return &pub, nil
}

// GoogleStart returns the authorize URL, OAuth state, and PKCE verifier for the gateway cookie.
func (s service) GoogleStart(context.Context) (string, string, string, error) {
	if !google.Configured(s.google) {
		return "", "", "", exception.Unavailable(exception.MsgUnavailable, nil)
	}

	state, err := google.RandomToken()
	if err != nil {
		return "", "", "", exception.Internal(err)
	}
	verifier, challenge, err := google.PKCE()
	if err != nil {
		return "", "", "", exception.Internal(err)
	}

	return google.AuthorizeURL(s.google, state, challenge), state, verifier, nil
}

// GoogleFinish links or creates a user from Google and opens a session.
func (s service) GoogleFinish(ctx context.Context, code, verifier string, meta models.ClientMeta) (*models.SessionResult, error) {
	if !google.Configured(s.google) {
		return nil, exception.Unavailable(exception.MsgUnavailable, nil)
	}
	if code == "" || len(code) > 2048 || verifier == "" || len(verifier) > 128 {
		return nil, exception.Unauthenticated(exception.MsgGoogleDenied, nil)
	}

	profile, err := google.Exchange(ctx, s.google, code, verifier)
	if errors.Is(err, google.ErrDenied) {
		return nil, exception.Unauthenticated(exception.MsgGoogleDenied, nil)
	}
	if errors.Is(err, google.ErrUnavailable) {
		return nil, exception.Unavailable(exception.MsgUnavailable, nil)
	}
	if err != nil {
		return nil, exception.Internal(err)
	}
	if !profile.Verified {
		return nil, exception.Unauthenticated(exception.MsgGoogleEmailUnverified, nil)
	}
	email, err := validator.NormalizeEmail(profile.Email)
	if err != nil {
		return nil, exception.Unauthenticated(exception.MsgGoogleDenied, nil)
	}

	if s.inviteOnly {
		if _, err := s.users.FindUserByEmail(ctx, email); errors.Is(err, constants.ErrNotFound) {
			if err := s.assertRegistrationAllowed(ctx, email, ""); err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}
	}

	if err := s.sessions.DeleteExpired(ctx, s.now()); err != nil {
		return nil, err
	}

	now := s.now()
	userID, err := token.NewID()
	if err != nil {
		return nil, exception.Internal(err)
	}
	identityID, err := token.NewID()
	if err != nil {
		return nil, exception.Internal(err)
	}

	handle, err := username.Allocate(ctx, s.users, email)
	if err != nil {
		return nil, err
	}

	user, err := s.users.SaveGoogleUser(ctx, models.User{
		ID:           userID,
		Email:        email,
		Username:     handle,
		PasswordHash: "",
		DisplayName:  google.DisplayName(profile.Name, email, constants.MaxDisplayNameRunes),
		AvatarURL:    profile.AvatarURL,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, identityID, constants.GoogleOAuthProvider, profile.Subject)
	if err != nil {
		return nil, err
	}

	if err := ensureAccountActive(user); err != nil {
		return nil, err
	}

	if _, mfaRequired, err := s.maybeBeginLoginMFA(ctx, user, meta.UserAgent, meta.ClientIP); err != nil {
		return nil, err
	} else if mfaRequired {
		return nil, exception.Unauthenticated(exception.MsgGoogleMFAUsePassword, nil)
	}

	return s.openSession(ctx, user, meta)
}

func registerUnavailable() error {
	return exception.InvalidInput(exception.MsgRegisterUnavailable, nil)
}

func (s service) assertRegistrationAllowed(ctx context.Context, email, inviteToken string) error {
	if !s.inviteOnly {
		return nil
	}

	n, err := s.users.CountUsers(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return nil
	}

	token := strings.TrimSpace(inviteToken)
	if token == "" {
		return exception.PermissionDenied(exception.MsgRegistrationClosed, nil)
	}
	if s.invites == nil {
		return exception.PermissionDenied(exception.MsgRegistrationClosed, nil)
	}

	row, err := s.invites.FindByToken(ctx, token)
	if err != nil {
		return exception.PermissionDenied(exception.MsgRegistrationClosed, nil)
	}

	now := s.now()
	if row.RedeemedAt != nil || !row.ExpiresAt.After(now) {
		return exception.PermissionDenied(exception.MsgRegistrationClosed, nil)
	}
	if row.Email != nil {
		bound := strings.TrimSpace(*row.Email)
		if bound != "" && !strings.EqualFold(bound, email) {
			return exception.PermissionDenied(exception.MsgRegistrationClosed, nil)
		}
	}

	return nil
}

func ensureAccountActive(user models.User) error {
	if user.DeletedAt != nil {
		return exception.Unauthenticated(exception.MsgAccountDeleted, nil)
	}
	if user.DisabledAt != nil {
		return exception.PermissionDenied(exception.MsgAccountDisabled, nil)
	}

	return nil
}

func (s service) openSession(ctx context.Context, user models.User, meta models.ClientMeta) (*models.SessionResult, error) {
	if err := ensureAccountActive(user); err != nil {
		return nil, err
	}
	now := s.now()
	sessionID, err := token.NewID()
	if err != nil {
		return nil, exception.Internal(err)
	}

	plain, hash, err := token.NewRefresh()
	if err != nil {
		return nil, exception.Internal(err)
	}
	refreshID, err := token.NewID()
	if err != nil {
		return nil, exception.Internal(err)
	}

	ua := meta.UserAgent
	if len(ua) > constants.MaxUserAgentLen {
		ua = ua[:constants.MaxUserAgentLen]
	}

	session := models.Session{
		ID:               sessionID,
		UserID:           user.ID,
		ExpiresAt:        now.Add(constants.RefreshTokenTTL),
		UserAgent:        ua,
		ClientIP:         meta.ClientIP,
		DeviceLabel:      meta.DeviceLabel,
		Platform:         meta.Platform,
		LocationLabel:    meta.LocationLabel,
		RefreshID:        refreshID,
		RefreshTokenHash: hash,
		LastUsedAt:       now,
		CreatedAt:        now,
	}
	refresh := models.RefreshToken{
		ID:        refreshID,
		SessionID: sessionID,
		TokenHash: hash,
		CreatedAt: now,
	}
	if err := s.sessions.InsertSession(ctx, session); err != nil {
		return nil, err
	}
	if err := s.refresh.Put(ctx, refresh, user.ID, session.ExpiresAt); err != nil {
		_ = s.sessions.RevokeSession(ctx, sessionID, now)
		return nil, err
	}

	access, expiresIn, err := jwt.Issue(s.secret, user.ID, sessionID, now, constants.AccessTokenTTL)
	if err != nil {
		return nil, exception.Internal(err)
	}

	s.recordActivity(ctx, models.UserActivityLogEntry{
		UserID:     user.ID,
		ActorID:    user.ID,
		Category:   constants.ActivityCategoryAuth,
		Action:     constants.ActivityUserLoggedIn,
		TargetType: constants.ActivityTargetSession,
		TargetID:   sessionID,
		Summary:    "Logged in",
		Metadata: map[string]any{
			"platform":     meta.Platform,
			"device_label": meta.DeviceLabel,
			"client_ip":    meta.ClientIP,
		},
		CreatedAt: now,
	})
	return &models.SessionResult{
		AccessToken:  access,
		ExpiresIn:    expiresIn,
		RefreshToken: plain,
		User:         user.Public(),
	}, nil
}

func (s service) refreshReuseResponse(ctx context.Context, sessionID string, usedAt *time.Time) (*models.SessionResult, error) {
	if usedAt != nil && s.now().Sub(*usedAt) < constants.RefreshReuseRevokeGrace {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	return s.reuse(ctx, sessionID)
}

func (s service) reuse(ctx context.Context, sessionID string) (*models.SessionResult, error) {
	if s.log != nil {
		s.log.Warn("refresh token reuse", "session_id", sessionID)
	}
	now := s.now()
	_ = s.refresh.RevokeForSession(ctx, sessionID)
	if err := s.sessions.RevokeSession(ctx, sessionID, now); err != nil {
		return nil, err
	}

	return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
}

func (s service) revokeAllUserSessions(ctx context.Context, userID string, now time.Time) error {
	ids, err := s.sessions.ListSessionIDsForUser(ctx, userID)
	if err != nil {
		return err
	}
	if err := s.sessions.RevokeAllSessionsForUser(ctx, userID, now); err != nil {
		return err
	}
	for _, id := range ids {
		_ = s.refresh.RevokeForSession(ctx, id)
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

func (s service) maybeBeginLoginMFA(ctx context.Context, user models.User, userAgent, clientIP string) (*models.PasswordAuthResult, bool, error) {
	if s.mfa == nil {
		return nil, false, nil
	}

	totpEnabled := true
	if _, err := s.users.GetMFATOTP(ctx, user.ID); errors.Is(err, constants.ErrNotFound) {
		totpEnabled = false
	} else if err != nil {
		return nil, false, err
	}

	waCreds, err := s.users.ListWebAuthnCredentials(ctx, user.ID)
	if err != nil {
		return nil, false, err
	}
	if !totpEnabled && len(waCreds) == 0 {
		return nil, false, nil
	}

	var methods []string
	var waOptsJSON string
	var waSessJSON string

	if len(waCreds) > 0 && s.webauthn != nil {
		methods = append(methods, "webauthn")
		waUser := mfa.WebAuthnUser{
			ID:          []byte(user.ID),
			Name:        user.Username,
			DisplayName: user.DisplayName,
			Credentials: toAuthWebAuthnCreds(waCreds),
		}
		assertion, sessData, err := s.webauthn.BeginLogin(waUser)
		if err != nil {
			return nil, false, exception.Internal(err)
		}
		optsBytes, err := json.Marshal(assertion)
		if err != nil {
			return nil, false, exception.Internal(err)
		}
		sessBytes, err := json.Marshal(sessData)
		if err != nil {
			return nil, false, exception.Internal(err)
		}
		waOptsJSON = string(optsBytes)
		waSessJSON = string(sessBytes)
	}
	if totpEnabled {
		methods = append(methods, "totp")
	}
	methods = append(methods, "recovery")

	challengeID, err := token.NewID()
	if err != nil {
		return nil, false, exception.Internal(err)
	}
	if err := s.mfa.PutLoginMFA(ctx, challengeID, models.LoginMFAPending{
		UserID:              user.ID,
		Email:               user.Email,
		Methods:             methods,
		WebAuthnSessionJSON: waSessJSON,
		UserAgent:           userAgent,
		ClientIP:            clientIP,
	}, constants.MFAChallengeTTL); err != nil {
		return nil, false, err
	}

	return &models.PasswordAuthResult{
		Challenge: &models.AuthChallenge{
			ChallengeID:         challengeID,
			MFARequired:         true,
			MFAMethods:          methods,
			WebAuthnOptionsJSON: waOptsJSON,
		},
	}, true, nil
}

func (s service) BeginLoginWebAuthn(ctx context.Context, challengeID string) (string, error) {
	if s.mfa == nil || s.webauthn == nil {
		return "", exception.Unavailable(exception.MsgUnavailable, nil)
	}
	challengeID = strings.TrimSpace(challengeID)
	if challengeID == "" {
		return "", exception.InvalidInput(exception.MsgInvalidInput, nil)
	}

	pending, err := s.mfa.GetLoginMFA(ctx, challengeID)
	if errors.Is(err, constants.ErrNotFound) {
		return "", exception.Unauthenticated(exception.MsgResetCodeInvalid, nil)
	}
	if err != nil {
		return "", err
	}

	user, err := s.users.FindUserByID(ctx, pending.UserID)
	if err != nil {
		return "", err
	}
	waCreds, err := s.users.ListWebAuthnCredentials(ctx, user.ID)
	if err != nil {
		return "", err
	}
	if len(waCreds) == 0 {
		return "", exception.InvalidInput(exception.MsgInvalidInput, nil)
	}

	waUser := mfa.WebAuthnUser{
		ID:          []byte(user.ID),
		Name:        user.Username,
		DisplayName: user.DisplayName,
		Credentials: toAuthWebAuthnCreds(waCreds),
	}
	assertion, sessData, err := s.webauthn.BeginLogin(waUser)
	if err != nil {
		return "", exception.Internal(err)
	}
	optsBytes, err := json.Marshal(assertion)
	if err != nil {
		return "", exception.Internal(err)
	}
	sessBytes, err := json.Marshal(sessData)
	if err != nil {
		return "", exception.Internal(err)
	}

	pending.WebAuthnSessionJSON = string(sessBytes)
	if err := s.mfa.UpdateLoginMFA(ctx, challengeID, pending, constants.MFAChallengeTTL); err != nil {
		return "", err
	}

	return string(optsBytes), nil
}

func (s service) VerifyLoginMFA(ctx context.Context, challengeID, method, code, assertionJSON string) (*models.SessionResult, error) {
	if s.mfa == nil {
		return nil, exception.Unavailable(exception.MsgUnavailable, nil)
	}
	challengeID = strings.TrimSpace(challengeID)
	method = strings.TrimSpace(method)
	if challengeID == "" || method == "" {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, nil)
	}

	pending, err := s.mfa.GetLoginMFA(ctx, challengeID)
	if errors.Is(err, constants.ErrNotFound) {
		return nil, exception.Unauthenticated(exception.MsgResetCodeInvalid, nil)
	}
	if err != nil {
		return nil, err
	}
	if pending.Attempts >= constants.MFAMaxAttempts {
		_ = s.mfa.DeleteLoginMFA(ctx, challengeID)
		return nil, exception.Unauthenticated(exception.MsgResetCodeInvalid, nil)
	}

	user, err := s.users.FindUserByID(ctx, pending.UserID)
	if err != nil {
		return nil, err
	}

	now := s.now()
	var verified bool

	switch method {
	case "totp":
		totpRec, err := s.users.GetMFATOTP(ctx, user.ID)
		if errors.Is(err, constants.ErrNotFound) {
			return nil, exception.InvalidInput(exception.MsgInvalidInput, nil)
		}
		if err != nil {
			return nil, err
		}
		secret, err := mfa.DecryptSecret(s.mfaKey, totpRec.SecretEncrypted)
		if err != nil {
			return nil, exception.Internal(err)
		}
		if step, ok := mfa.VerifyTOTP(secret, code, now, totpRec.LastUsedStep); ok {
			if err := s.users.UpdateMFATOTPLastUsedStep(ctx, user.ID, step, now); err != nil {
				return nil, err
			}
			verified = true
		}
	case "recovery":
		if strings.TrimSpace(code) != "" {
			hash := mfa.HashRecoveryCode(s.secret, code)
			ok, err := s.users.ConsumeRecoveryCode(ctx, user.ID, hash, now)
			if err != nil {
				return nil, err
			}
			verified = ok
		}
	case "webauthn":
		if s.webauthn == nil || pending.WebAuthnSessionJSON == "" || strings.TrimSpace(assertionJSON) == "" {
			return nil, exception.InvalidInput(exception.MsgInvalidInput, nil)
		}
		var sessData webauthn.SessionData
		if err := json.Unmarshal([]byte(pending.WebAuthnSessionJSON), &sessData); err != nil {
			return nil, exception.Internal(err)
		}
		parsed, err := protocol.ParseCredentialRequestResponseBody(strings.NewReader(assertionJSON))
		if err != nil {
			return nil, exception.InvalidInput(exception.MsgInvalidInput, nil)
		}
		waCreds, err := s.users.ListWebAuthnCredentials(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		waUser := mfa.WebAuthnUser{
			ID:          []byte(user.ID),
			Name:        user.Username,
			DisplayName: user.DisplayName,
			Credentials: toAuthWebAuthnCreds(waCreds),
		}
		cred, err := s.webauthn.ValidateLogin(waUser, sessData, parsed)
		if err == nil && cred != nil {
			if err := s.users.UpdateWebAuthnCredentialUsage(
				ctx,
				cred.ID,
				cred.Authenticator.SignCount,
				cred.Authenticator.CloneWarning,
				cred.Flags.BackupState,
				now,
			); err != nil {
				return nil, err
			}
			verified = true
		}
	default:
		return nil, exception.InvalidInput(exception.MsgInvalidInput, nil)
	}

	if !verified {
		pending.Attempts++
		if pending.Attempts >= constants.MFAMaxAttempts {
			_ = s.mfa.DeleteLoginMFA(ctx, challengeID)
		} else {
			_ = s.mfa.UpdateLoginMFA(ctx, challengeID, pending, constants.MFAChallengeTTL)
		}
		return nil, exception.Unauthenticated(exception.MsgResetCodeInvalid, nil)
	}

	sess, err := s.openSession(ctx, user, models.ClientMeta{UserAgent: pending.UserAgent, ClientIP: pending.ClientIP})
	if err != nil {
		return nil, err
	}
	_ = s.mfa.DeleteLoginMFA(ctx, challengeID)

	return sess, nil
}

func toAuthWebAuthnCreds(creds []models.WebAuthnCredential) []webauthn.Credential {
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
