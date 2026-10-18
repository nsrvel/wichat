// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

// Package authPb maps auth domain models to auth v1 protobuf messages.
package authPb

import (
	"github.com/wichat/wichat/backend/ms-user/internal/models"
	authv1 "github.com/wichat/wichat/backend/wi-shared/models/gen/auth/v1"
)

// Meta builds client hints from gRPC request fields.
func Meta(userAgent, clientIP string) models.ClientMeta {
	return models.ClientMeta{UserAgent: userAgent, ClientIP: clientIP}
}

// Session maps a service session result to protobuf.
func Session(out *models.SessionResult) *authv1.Session {
	if out == nil {
		return nil
	}

	return &authv1.Session{
		AccessToken:  out.AccessToken,
		ExpiresIn:    out.ExpiresIn,
		RefreshToken: out.RefreshToken,
		User:         User(out.User),
	}
}

// User maps a public user to protobuf.
func User(user models.PublicUser) *authv1.User {
	return &authv1.User{
		Id:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
	}
}
