package service

import (
	"context"
	"errors"
	"strings"
	"time"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	userservice "github.com/wichat/wichat/backend/ms-user/internal/user/service"
	authRedis "github.com/wichat/wichat/backend/ms-user/pkg/auth-redis"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/jwt"
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
	RegisterGuest(ctx context.Context, in models.GuestRegisterInput) (*models.SessionResult, error)
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
}

type service struct {
	users     userrepo.PgRepository
	sessions  authrepo.PgRepository
	workspace userservice.Service
	reset     authrepo.PasswordResetStore
	emailAuth authrepo.EmailAuthStore
	notify    Notifier
	log       logger.Logger
	secret    []byte
	now       func() time.Time
	dummy     string
	google    google.Config
}

// NewService builds the auth use cases. secret is the HMAC key and must be at least 32 bytes.
func NewService(
	users userrepo.PgRepository,
	sessions authrepo.PgRepository,
	secret []byte,
	log logger.Logger,
	googleCfg google.Config,
	reset authrepo.PasswordResetStore,
	emailAuth authrepo.EmailAuthStore,
	notify Notifier,
	workspace userservice.Service,
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

	return service{
		users:     users,
		sessions:  sessions,
		workspace: workspace,
		reset:     reset,
		emailAuth: emailAuth,
		notify:    notify,
		log:       log,
		secret:    copied,
		now:       func() time.Time { return time.Now().UTC() },
		dummy:     dummy,
		google:    google.Prepare(googleCfg),
	}, nil
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
	email, name, birth, err := validator.RegisterInput(in)
	if err != nil {
		return nil, err
	}

	if err := s.sessions.DeleteExpired(ctx, s.now()); err != nil {
		return nil, err
	}

	if _, err := s.users.FindUserByEmail(ctx, email); err == nil {
		return nil, exception.Conflict(exception.MsgEmailTaken, nil)
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
		PasswordHash: hash,
		DisplayName:  name,
		BirthDate:    birthStr,
		Locale:       strings.TrimSpace(in.Locale),
		CodeHMAC:     authRedis.OTPCodeHMAC(s.secret, code),
		UserAgent:    in.UserAgent,
		ClientIP:     in.ClientIP,
	}
	if err := s.emailAuth.PutRegisterPending(ctx, ch.ChallengeID, pending, constants.EmailAuthOTPTTL); err != nil {
		return nil, err
	}

	if err := s.sendEmailCode(ctx, email, name, in.Locale, "sign_up", code); err != nil {
		_ = s.emailAuth.DeleteRegisterPending(ctx, ch.ChallengeID)
		return nil, err
	}
	if err := s.markEmailCodeSent(ctx, ch, email); err != nil {
		return nil, err
	}

	return &models.PasswordAuthResult{Challenge: ch}, nil
}

func (s service) createUserFromPending(ctx context.Context, pending models.RegisterPending) (models.User, error) {
	if _, err := s.users.FindUserByEmail(ctx, pending.Email); err == nil {
		return models.User{}, exception.Conflict(exception.MsgEmailTaken, nil)
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

	handle := ""
	assignUsername, err := s.workspace.UsernameEnabledForNewUsers(ctx)
	if err != nil {
		return models.User{}, err
	}
	if assignUsername {
		handle, err = username.Allocate(ctx, s.users, pending.Email)
		if err != nil {
			return models.User{}, err
		}
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
			return models.User{}, exception.Conflict(exception.MsgEmailTaken, nil)
		}

		return models.User{}, err
	}

	return user, nil
}

// RegisterGuest creates a guest account and opens a session.
func (s service) RegisterGuest(ctx context.Context, in models.GuestRegisterInput) (*models.SessionResult, error) {
	if err := s.sessions.DeleteExpired(ctx, s.now()); err != nil {
		return nil, err
	}

	user, err := s.workspace.CreateGuest(ctx, in)
	if err != nil {
		return nil, err
	}

	return s.openSession(ctx, user, in.ClientMeta)
}

// Login checks login/password and starts email OTP (no session until VerifyLogin).
func (s service) Login(ctx context.Context, in models.LoginInput) (*models.PasswordAuthResult, error) {
	login := strings.TrimSpace(in.Login)
	if login == "" {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "login"})
	}
	if !password.ValidLength(in.Password) {
		return nil, exception.InvalidInput(exception.MsgInvalidInput, map[string]any{"field": "password"})
	}

	if err := s.sessions.DeleteExpired(ctx, s.now()); err != nil {
		return nil, err
	}

	user, findErr := s.users.FindUserByLogin(ctx, login)
	hash := s.dummy
	if findErr == nil && user.PasswordHash != "" {
		hash = user.PasswordHash
	}
	if !password.Verify(hash, in.Password) || findErr != nil || user.PasswordHash == "" {
		return nil, exception.Unauthenticated(exception.MsgInvalidCredentials, nil)
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

func (s service) markEmailCodeSent(ctx context.Context, ch *models.AuthChallenge, email string) error {
	wait, err := s.emailAuth.RecordEmailCodeSent(ctx, strings.TrimSpace(email))
	if err != nil {
		return err
	}
	ch.ResendCooldownSeconds = resendCooldownInt32(wait)

	return nil
}

func (s service) resendEmailCode(
	ctx context.Context,
	email, displayName, locale, purpose string,
	update func(codeHMAC string) error,
) (int32, error) {
	if err := s.requireEmailAuth(); err != nil {
		return 0, err
	}
	email = strings.TrimSpace(email)
	if err := s.emailAuth.EnsureCanSendEmailCode(ctx, email); err != nil {
		if retry, ok := models.ResetRateLimitRetryAfter(err); ok {
			return resendCooldownInt32(retry), resetRateLimited(retry)
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

	wait, err := s.emailAuth.RecordEmailCodeSent(ctx, email)
	if err != nil {
		return 0, err
	}

	return resendCooldownInt32(wait), nil
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

	if err := s.emailAuth.DeleteRegisterPending(ctx, challengeID); err != nil {
		return nil, err
	}
	if err := s.emailAuth.ClearEmailCodePenalty(ctx, pending.Email); err != nil {
		return nil, err
	}

	user, err := s.createUserFromPending(ctx, pending)
	if err != nil {
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

	if err := s.emailAuth.DeleteLoginChallenge(ctx, challengeID); err != nil {
		return nil, err
	}
	if err := s.emailAuth.ClearEmailCodePenalty(ctx, pending.Email); err != nil {
		return nil, err
	}

	user, err := s.users.FindUserByID(ctx, pending.UserID)
	if err != nil {
		return nil, err
	}

	return s.openSession(ctx, user, models.ClientMeta{UserAgent: pending.UserAgent, ClientIP: pending.ClientIP})
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

	return s.resendEmailCode(ctx, pending.Email, pending.DisplayName, pending.Locale, "sign_up", func(hmac string) error {
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

	return s.resendEmailCode(ctx, pending.Email, pending.DisplayName, pending.Locale, "sign_in", func(hmac string) error {
		return s.emailAuth.UpdateLoginCode(ctx, challengeID, hmac, constants.EmailAuthOTPTTL)
	})
}

// RequestPasswordReset emails a reset code when the account exists (silent otherwise).
func (s service) RequestPasswordReset(ctx context.Context, emailRaw, locale string) (int, error) {
	email, err := validator.NormalizeEmail(emailRaw)
	if err != nil {
		return 0, err
	}

	user, err := s.users.FindUserByEmail(ctx, email)
	if errors.Is(err, constants.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if user.PasswordHash == "" {
		return 0, nil
	}

	if err := s.requirePasswordReset(); err != nil {
		return 0, err
	}

	if err := s.reset.EnsureCanSendResetCode(ctx, email); err != nil {
		if retry, ok := models.ResetRateLimitRetryAfter(err); ok {
			return resendCooldownSecs(retry), resetRateLimited(retry)
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

	wait, err := s.reset.RecordResetCodeSent(ctx, email)
	if err != nil {
		return 0, err
	}

	return resendCooldownSecs(wait), nil
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

	if err := s.reset.DeleteOTP(ctx, email); err != nil {
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

	userID, email, err := s.reset.ConsumeResetToken(ctx, resetToken)
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

	if err := s.sessions.RevokeAllSessionsForUser(ctx, userID, now); err != nil {
		return err
	}

	_ = email // token record carries email for auditing later

	return nil
}

// Refresh rotates the refresh token and issues a new access JWT for the same session id.
func (s service) Refresh(ctx context.Context, refreshToken string, _ models.ClientMeta) (*models.SessionResult, error) {
	if refreshToken == "" {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	if err := s.sessions.DeleteExpired(ctx, s.now()); err != nil {
		return nil, err
	}

	rec, err := s.sessions.FindRefreshByHash(ctx, token.HashRefresh(refreshToken))
	if errors.Is(err, constants.ErrNotFound) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return nil, err
	}

	sess, err := s.sessions.FindSession(ctx, rec.SessionID)
	if errors.Is(err, constants.ErrNotFound) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return nil, err
	}

	if sess.RevokedAt != nil || !sess.ExpiresAt.After(s.now()) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if rec.UsedAt != nil {
		return s.reuse(ctx, sess.ID)
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
	err = s.sessions.MarkRefreshUsed(ctx, token.HashRefresh(refreshToken), nextHash, nextID, sess.ID, now.Add(constants.RefreshTokenTTL), now)
	if errors.Is(err, constants.ErrAlreadyUsed) {
		return s.reuse(ctx, sess.ID)
	}
	if errors.Is(err, constants.ErrNotFound) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return nil, err
	}

	user, err := s.users.FindUserByID(ctx, sess.UserID)
	if errors.Is(err, constants.ErrNotFound) {
		return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
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

	rec, err := s.sessions.FindRefreshByHash(ctx, token.HashRefresh(refreshToken))
	if errors.Is(err, constants.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	return s.sessions.RevokeSession(ctx, rec.SessionID, s.now())
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

	user, err := s.users.SaveGoogleUser(ctx, models.User{
		ID:           userID,
		Email:        email,
		PasswordHash: "",
		DisplayName:  google.DisplayName(profile.Name, email, constants.MaxDisplayNameRunes),
		CreatedAt:    now,
		UpdatedAt:    now,
	}, identityID, constants.GoogleOAuthProvider, profile.Subject)
	if err != nil {
		return nil, err
	}

	return s.openSession(ctx, user, meta)
}

func (s service) openSession(ctx context.Context, user models.User, meta models.ClientMeta) (*models.SessionResult, error) {
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
		ID:         sessionID,
		UserID:     user.ID,
		ExpiresAt:  now.Add(constants.RefreshTokenTTL),
		UserAgent:  ua,
		ClientIP:   meta.ClientIP,
		LastUsedAt: now,
		CreatedAt:  now,
	}
	refresh := models.RefreshToken{
		ID:        refreshID,
		SessionID: sessionID,
		TokenHash: hash,
		CreatedAt: now,
	}
	if err := s.sessions.InsertSession(ctx, session, refresh); err != nil {
		return nil, err
	}

	access, expiresIn, err := jwt.Issue(s.secret, user.ID, sessionID, now, constants.AccessTokenTTL)
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

func (s service) reuse(ctx context.Context, sessionID string) (*models.SessionResult, error) {
	if s.log != nil {
		s.log.Warn("refresh token reuse", "session_id", sessionID)
	}
	if err := s.sessions.RevokeSession(ctx, sessionID, s.now()); err != nil {
		return nil, err
	}

	return nil, exception.Unauthenticated(exception.MsgUnauthorized, nil)
}

func resendCooldownSecs(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	secs := int((d + time.Second - 1) / time.Second)
	if secs < 1 {
		return 1
	}

	return secs
}

func resendCooldownInt32(d time.Duration) int32 {
	secs := resendCooldownSecs(d)
	if secs > 1_000_000 {
		return 1_000_000
	}

	return int32(secs) //nolint:gosec // G115 — bounded above
}

func resetRateLimited(retry time.Duration) error {
	secs := resendCooldownSecs(retry)
	params := map[string]any{"retry_after_seconds": secs}
	if secs < 1 {
		params["retry_after_seconds"] = 1
	}

	return exception.RateLimited(exception.MsgResetRateLimited, params)
}
