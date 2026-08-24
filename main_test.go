package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testServer(maxSpans int) *Server {
	return &Server{
		store:       NewTraceStore(maxSpans),
		metrics:     &Metrics{},
		maxBody:     4096,
		maxAttrs:    4,
		maxAttrSize: 64,
	}
}

func spanPayload(traceID, spanID string) []byte {
	payload, _ := json.Marshal(Span{
		TraceID:    traceID,
		SpanID:     spanID,
		Service:    "api",
		Operation:  "GET_users",
		DurationMS: 12,
		Timestamp:  time.Unix(1_700_000_000, 0).UTC(),
		Attributes: map[string]string{"region": "test"},
	})
	return payload
}

func TestHealthReadinessAndMetrics(t *testing.T) {
	h := testServer(10).Handler()
	for _, path := range []string{"/healthz", "/readyz", "/metrics"} {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", path, rr.Code)
		}
	}
}

func TestIngestAndQueryTrace(t *testing.T) {
	s := testServer(10)
	h := s.Handler()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/spans", bytes.NewReader(spanPayload("trace-1", "span-1"))))
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d: %s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/api/v1/traces/trace-1", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if s.metrics.received.Load() != 1 || s.store.Count() != 1 {
		t.Fatal("expected one recorded span")
	}
}

func TestValidationRejectsMalformedSpan(t *testing.T) {
	s := testServer(10)
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/spans", bytes.NewBufferString(`{"trace_id":"bad id"}`)))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	if s.metrics.rejected.Load() != 1 {
		t.Fatal("expected rejected metric")
	}
}

func TestBoundedStoreEvictsOldestSpan(t *testing.T) {
	s := testServer(2)
	h := s.Handler()
	for i, id := range []string{"one", "two", "three"} {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/spans", bytes.NewReader(spanPayload("trace-"+id, "span-"+id))))
		if rr.Code != http.StatusAccepted {
			t.Fatalf("ingest %d failed: %d", i, rr.Code)
		}
	}
	if s.store.Count() != 2 || s.metrics.evicted.Load() != 1 {
		t.Fatalf("expected bounded store and one eviction")
	}
	if got := s.store.ByTrace("trace-one"); len(got) != 0 {
		t.Fatal("oldest span should have been evicted")
	}
}

func TestBearerBoundary(t *testing.T) {
	s := testServer(10)
	s.apiToken = "0123456789abcdef"
	h := s.Handler()

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/api/v1/spans", bytes.NewReader(spanPayload("trace-1", "span-1"))))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/spans", bytes.NewReader(spanPayload("trace-1", "span-1")))
	req.Header.Set("Authorization", "Bearer 0123456789abcdef")
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rr.Code)
	}
}
