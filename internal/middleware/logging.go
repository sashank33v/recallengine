package middleware

import (
	"expvar"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"
)

var (
	requestsTotal      = expvar.NewInt("requests_total")
	requestsByStatus   = expvar.NewMap("requests_by_status")
	requestBYRoute     = expvar.NewMap("requests_by_route")
	requestMillisTotal = expvar.NewInt("request_duration_ms_total")
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Flush() {
	f, ok := w.ResponseWriter.(http.Flusher)
	if ok {
		f.Flush()
	}
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(sw, r)
		route := r.Pattern
		if route == "" {
			route = "unmatched"
			requestsTotal.Add(1)
			requestsByStatus.Add(strconv.Itoa(sw.status), 1)
			requestBYRoute.Add(route, 1)
			requestMillisTotal.Add(int64(time.Since(start).Milliseconds()))
		}
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"duration_ms", time.Since(start).Microseconds(),
		)
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				slog.Error("panic in handler",
					"error", err,
					"path", r.URL.Path,
					"stack", string(debug.Stack()),
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"something went wrong"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
