// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2026 Garam

// Package jwt issues and validates short-lived HS256 access tokens for auth sessions.
package jwt

import (
	"errors"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

type accessClaims struct {
	SID string `json:"sid"`
	jwtlib.RegisteredClaims
}

// Issue signs an access token bound to userID and sessionID (sid claim).
func Issue(secret []byte, userID, sessionID string, now time.Time, ttl time.Duration) (string, int32, error) {
	secs := int64(ttl / time.Second)
	if secs < 0 || secs > int64(^uint32(0)>>1) {
		return "", 0, errors.New("jwt: ttl out of range")
	}
	expiresIn := int32(secs)
	claims := accessClaims{
		SID: sessionID,
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return "", 0, err
	}

	return signed, expiresIn, nil
}

// Parse validates raw and returns userID (sub) and sessionID (sid).
func Parse(secret []byte, raw string) (userID, sessionID string, err error) {
	var claims accessClaims
	parsed, err := jwtlib.ParseWithClaims(raw, &claims, func(token *jwtlib.Token) (any, error) {
		return secret, nil
	}, jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()}))
	if err != nil || !parsed.Valid {
		return "", "", err
	}
	if claims.Subject == "" || claims.SID == "" {
		return "", "", errors.New("jwt: missing claims")
	}

	return claims.Subject, claims.SID, nil
}
