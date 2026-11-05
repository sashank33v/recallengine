package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

type window struct {
	start time.Time
	count int
}

type RateLimiter struct {
	mu     sync.Mutex
	limit  int
	period time.Duration
	users  map[int64]*window
}

func NewRateLimiter(limit int, period time.Duration) *RateLimiter {
	return &RateLimiter{limit: limit, period: period, users: map[int64]*window{}}
}

func (l *RateLimiter) allow(userID int64) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	if len(l.users) > 1000 {
		for id, w := range l.users {
			if now.Sub(w.start) >= l.period {
				delete(l.users, id)
			}
		}
	}
	w, found := l.users[userID]
	if !found || now.Sub(w.start) >= l.period {
		l.users[userID] = &window{start: now, count: 1}
		return true, 0
	}
	if w.count >= l.limit {
		return false, l.period - now.Sub(w.start)
	}
	w.count++
	return true, 0
}

func (l *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := UserIDFromContext(r.Context())
		if !ok {
			writeUnauthorized(w)
			return
		}
		allowed, wait := l.allow(userID)
		if !allowed {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", strconv.Itoa(int(wait.Seconds())+1))
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"too many requests, please wait for a moment"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
