package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestStartRequiresSpanName(t *testing.T) {
	_, _, err := Start(context.Background(), otel.Tracer("test"), "")
	if err == nil {
		t.Fatal("expected empty span name to be rejected")
	}
}

func TestHTTPHandlerWrapsHandler(t *testing.T) {
	handler := HTTPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), "test.request")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", recorder.Code)
	}
}
