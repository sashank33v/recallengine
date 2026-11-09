package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoverReturns500(t *testing.T) {
	handler := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", rec.Code)
	}
}

func TestLoggingKeepsStatus(t *testing.T) {
	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", rec.Code)
	}
}

func TestStatusWriterSupportsFlush(t *testing.T) {
	var w http.ResponseWriter = &statusWriter{ResponseWriter: httptest.NewRecorder(), status: http.StatusOK}

	if _, ok := w.(http.Flusher); !ok {
		t.Error("statusWriter must support Flush, or streaming breaks")
	}
}
func TestLoggingCountsRequests(t *testing.T) {
	before := requestsTotal.Value()

	handler := Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))
	if requestsTotal.Value() != before+1 {
		t.Error("the request was not counted")
	}
	if requestsByStatus.Get("418") == nil {
		t.Error("status 418 was not counted")
	}
}
