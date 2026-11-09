package handler

import (
	"crypto/subtle"
	"expvar"
	"net/http"
	"strings"
)

func Metrics(token string) http.Handler {
	counters := expvar.Handler()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token == "" {
			http.NotFound(w, r)
			return
		}
		given := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if subtle.ConstantTimeCompare([]byte(given), []byte(token)) != 1 {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		counters.ServeHTTP(w, r)
	})
}
