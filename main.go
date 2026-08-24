package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const serviceName = "sky-observe"

var idPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{1,128}$`)

type Span struct {
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	ParentID   string            `json:"parent_id,omitempty"`
	Service    string            `json:"service"`
	Operation  string            `json:"operation"`
	DurationMS int64             `json:"duration_ms"`
	Timestamp  time.Time         `json:"timestamp"`
	Status     string            `json:"status,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type TraceStore struct {
	mu       sync.RWMutex
	spans    []Span
	maxSpans int
}

func NewTraceStore(maxSpans int) *TraceStore {
	if maxSpans < 1 {
		panic("maxSpans must be positive")
	}
	return &TraceStore{maxSpans: maxSpans}
}

func (s *TraceStore) Add(span Span) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.spans) == s.maxSpans {
		copy(s.spans, s.spans[1:])
		s.spans[len(s.spans)-1] = span
		return
	}
	s.spans = append(s.spans, span)
}

func (s *TraceStore) ByTrace(traceID string) []Span {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Span, 0)
	for _, span := range s.spans {
		if span.TraceID == traceID {
			out = append(out, span)
		}
	}
	return out
}

func (s *TraceStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.spans)
}

type Metrics struct {
	received atomic.Uint64
	rejected atomic.Uint64
	evicted  atomic.Uint64
}

type Server struct {
	store       *TraceStore
	metrics     *Metrics
	apiToken    string
	maxBody     int64
	maxAttrs    int
	maxAttrSize int
}

func envInt(name string, fallback, min, max int) int {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < min || value > max {
		log.Fatalf("%s must be an integer between %d and %d", name, min, max)
	}
	return value
}

func validSpan(span Span, maxAttrs, maxAttrSize int) error {
	if !idPattern.MatchString(span.TraceID) || !idPattern.MatchString(span.SpanID) {
		return errors.New("trace_id and span_id must be 1-128 safe characters")
	}
	if span.ParentID != "" && !idPattern.MatchString(span.ParentID) {
		return errors.New("invalid parent_id")
	}
	if !idPattern.MatchString(span.Service) || !idPattern.MatchString(span.Operation) {
		return errors.New("service and operation are required safe identifiers")
	}
	if span.DurationMS < 0 || span.DurationMS > 24*60*60*1000 {
		return errors.New("duration_ms is out of range")
	}
	if span.Timestamp.IsZero() {
		return errors.New("timestamp is required")
	}
	if len(span.Attributes) > maxAttrs {
		return errors.New("too many attributes")
	}
	for key, value := range span.Attributes {
		if !idPattern.MatchString(key) || len(value) > maxAttrSize {
			return errors.New("invalid attribute")
		}
	}
	return nil
}

func (s *Server) authorized(r *http.Request) bool {
	if s.apiToken == "" {
		return true
	}
	expected := "Bearer " + s.apiToken
	supplied := r.Header.Get("Authorization")
	return len(expected) == len(supplied) && subtle.ConstantTimeCompare([]byte(expected), []byte(supplied)) == 1
}

func writeJSON(w http.ResponseWriter, code int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(value)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "healthy", "service": serviceName})
}

func (s *Server) handleReady(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ready", "service": serviceName, "stored_spans": s.store.Count()})
}

func (s *Server) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service":      serviceName,
		"stored_spans": s.store.Count(),
		"received":     s.metrics.received.Load(),
		"rejected":     s.metrics.rejected.Load(),
		"evicted":      s.metrics.evicted.Load(),
	})
}

func (s *Server) handleSpans(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		s.metrics.rejected.Add(1)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	body := http.MaxBytesReader(w, r.Body, s.maxBody)
	defer body.Close()
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	var span Span
	if err := decoder.Decode(&span); err != nil {
		s.metrics.rejected.Add(1)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid span payload"})
		return
	}
	if err := validSpan(span, s.maxAttrs, s.maxAttrSize); err != nil {
		s.metrics.rejected.Add(1)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	before := s.store.Count()
	s.store.Add(span)
	if before == s.store.maxSpans {
		s.metrics.evicted.Add(1)
	}
	s.metrics.received.Add(1)
	writeJSON(w, http.StatusAccepted, map[string]any{"status": "recorded", "trace_id": span.TraceID, "span_id": span.SpanID})
}

func (s *Server) handleTrace(w http.ResponseWriter, r *http.Request) {
	if !s.authorized(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	traceID := strings.TrimPrefix(r.URL.Path, "/api/v1/traces/")
	if !idPattern.MatchString(traceID) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid trace id"})
		return
	}
	spans := s.store.ByTrace(traceID)
	if len(spans) == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "trace not found"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"trace_id": traceID, "spans": spans, "count": len(spans)})
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/readyz", s.handleReady)
	mux.HandleFunc("/metrics", s.handleMetrics)
	mux.HandleFunc("/api/v1/spans", s.handleSpans)
	mux.HandleFunc("/api/v1/traces/", s.handleTrace)
	return mux
}

func main() {
	maxSpans := envInt("OBSERVE_MAX_SPANS", 10_000, 100, 1_000_000)
	maxBody := envInt("OBSERVE_MAX_BODY_BYTES", 64*1024, 1024, 1024*1024)
	apiToken := os.Getenv("OBSERVE_API_TOKEN")
	if apiToken != "" && len(apiToken) < 16 {
		log.Fatal("OBSERVE_API_TOKEN must contain at least 16 characters when configured")
	}
	server := &Server{
		store:       NewTraceStore(maxSpans),
		metrics:     &Metrics{},
		apiToken:    apiToken,
		maxBody:     int64(maxBody),
		maxAttrs:    64,
		maxAttrSize: 2048,
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		log.Printf("%s listening on %s", serviceName, addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
