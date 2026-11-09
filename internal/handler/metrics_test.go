package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsAccess(t *testing.T) {
	tests := []struct {
		name   string
		token  string
		header string
		want   int
	}{
		{"disabled without a token", "", "", http.StatusNotFound},
		{"missing header", "secret", "", http.StatusUnauthorized},
		{"wrong token", "secret", "Bearer nope", http.StatusUnauthorized},
		{"right token", "secret", "Bearer secret", http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()

			Metrics(tc.token).ServeHTTP(rec, req)

			if rec.Code != tc.want {
				t.Errorf("status: got %d, want %d", rec.Code, tc.want)
			}
		})
	}
}
