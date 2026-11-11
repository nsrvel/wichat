// Package cookie holds HttpOnly auth cookies for the api-gateway.
package cookie

import (
	"net/http"
	"strings"
)

const (
	RefreshName   = "wichat_refresh"
	RefreshPath   = "/api/v1/auth"
	refreshMaxAge = 30 * 24 * 60 * 60

	OAuthName       = "wichat_oauth"
	OAuthReturnName = "wichat_oauth_return"
	OAuthPath       = "/api" // covers /api/v1/auth/google and /api/auth/callback/google
	oauthMaxAge     = 10 * 60
)

func SetRefresh(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshName,
		Value:    token,
		Path:     RefreshPath,
		MaxAge:   refreshMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearRefresh(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshName,
		Value:    "",
		Path:     RefreshPath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func RefreshToken(r *http.Request) string {
	return value(r, RefreshName)
}

func SetOAuth(w http.ResponseWriter, state, verifier string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     OAuthName,
		Value:    state + "." + verifier,
		Path:     OAuthPath,
		MaxAge:   oauthMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func SetOAuthReturn(w http.ResponseWriter, returnPath string, secure bool) {
	returnPath = strings.TrimSpace(returnPath)
	if returnPath == "" || returnPath == "/" {
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     OAuthReturnName,
		Value:    returnPath,
		Path:     OAuthPath,
		MaxAge:   oauthMaxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func OAuthReturn(r *http.Request) string {
	return strings.TrimSpace(value(r, OAuthReturnName))
}

func ClearOAuth(w http.ResponseWriter, secure bool) {
	for _, name := range []string{OAuthName, OAuthReturnName} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     OAuthPath,
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   secure,
			SameSite: http.SameSiteLaxMode,
		})
	}
}

func OAuthStateVerifier(r *http.Request) (state, verifier string) {
	state, verifier, ok := strings.Cut(value(r, OAuthName), ".")
	if !ok || state == "" || verifier == "" {
		return "", ""
	}

	return state, verifier
}

func value(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}

	return c.Value
}
