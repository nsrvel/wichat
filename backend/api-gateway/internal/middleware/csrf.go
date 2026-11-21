package middleware

import (
	"net/http"

	"github.com/wichat/wichat/backend/wi-shared/exception"
	responsehttp "github.com/wichat/wichat/backend/wi-shared/response/http"
)

// CrossOriginStateChange rejects browser POST/PUT/PATCH/DELETE when Origin is present but not allowed.
// Same-origin and non-browser clients (no Origin) pass through. Pair with CORS + SameSite cookies.
func CrossOriginStateChange(allowedOrigin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			origin := r.Header.Get("Origin")
			if origin != "" && !corsOriginAllowed(allowedOrigin, origin) {
				responsehttp.WriteError(w, r, exception.PermissionDenied(exception.MsgUnauthorized, nil))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
