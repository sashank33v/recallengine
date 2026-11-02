package middleware

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/sashank33v/recallengine/internal/service"
)

type contextKey string

const userIDkey contextKey = "userID"

func Auth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token := strings.TrimPrefix(header, "Bearer ")
			if header == "" || token == header {
				writeUnauthorized(w)
				return
			}

			userID, err := authService.Authenticate(r.Context(), token)
			if errors.Is(err, sql.ErrNoRows) {
				writeUnauthorized(w)
				return
			}
			if err != nil {
				log.Println("auth error:", err)
				http.Error(w, `{"error":"something went wrong"}`, http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), userIDkey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDkey).(int64)
	return userID, ok
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"unauthorized"}`))
}
