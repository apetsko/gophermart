package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/apetsko/gophermart/internal/auth"
	"github.com/apetsko/gophermart/internal/logging"
)

func AuthMiddleware(secret string, logger *logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := auth.CookieGetUserID(r, secret)
			if err != nil {
				logger.Errorw("Cookie GetUserID", "error", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			r.Header.Set("UserID", strconv.Itoa(*userID))
			logger.Debug(fmt.Sprintf("userID: %d", *userID))
			next.ServeHTTP(w, r)
		})
	}
}
