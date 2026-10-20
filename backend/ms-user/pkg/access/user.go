package access

import (
	"context"
	"errors"
	"time"

	authrepo "github.com/wichat/wichat/backend/ms-user/internal/auth/repository"
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userrepo "github.com/wichat/wichat/backend/ms-user/internal/user/repository"
	"github.com/wichat/wichat/backend/ms-user/pkg/constants"
	"github.com/wichat/wichat/backend/ms-user/pkg/jwt"
	"github.com/wichat/wichat/backend/wi-shared/exception"
)

// UserFromAccessToken loads the account for a valid access token and live session.
func UserFromAccessToken(
	ctx context.Context,
	accessToken string,
	secret []byte,
	sessions authrepo.PgRepository,
	users userrepo.PgRepository,
	now time.Time,
) (models.User, error) {
	if accessToken == "" {
		return models.User{}, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	userID, sessionID, err := jwt.Parse(secret, accessToken)
	if err != nil {
		return models.User{}, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	sess, err := sessions.FindSession(ctx, sessionID)
	if errors.Is(err, constants.ErrNotFound) {
		return models.User{}, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return models.User{}, err
	}

	if sess.UserID != userID || sess.RevokedAt != nil || !sess.ExpiresAt.After(now) {
		return models.User{}, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}

	user, err := users.FindUserByID(ctx, userID)
	if errors.Is(err, constants.ErrNotFound) {
		return models.User{}, exception.Unauthenticated(exception.MsgUnauthorized, nil)
	}
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}
