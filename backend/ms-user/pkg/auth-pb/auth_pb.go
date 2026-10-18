// Package authPb maps auth domain models to auth v1 protobuf messages.
package authPb

import (
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	userv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/user/v1"
)

// Meta builds client hints from gRPC request fields.
func Meta(userAgent, clientIP string) models.ClientMeta {
	return models.ClientMeta{UserAgent: userAgent, ClientIP: clientIP}
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
	}
}
