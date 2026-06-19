// Distributed-Tracing-Setup: Collects and stores distributed trace spans
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Span struct {
	TraceID  string `json:"trace_id"`
	SpanID   string `json:"span_id"`
	Service  string `json:"service"`
	Duration int64  `json:"duration_ms"`
}

var spans []Span

func handleProcess(w http.ResponseWriter, r *http.Request) {
	var s Span
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	spans = append(spans, s)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "recorded", "span_id": s.SpanID, "total_spans": len(spans)})
}


func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"service":   "Distributed-Tracing-Setup",
		"timestamp": time.Now().Unix(),
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/api/v1/process", handleProcess)
	log.Printf("Distributed-Tracing-Setup running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
