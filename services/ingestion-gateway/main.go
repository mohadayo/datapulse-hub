package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Event struct {
	ID         string                 `json:"id"`
	PipelineID string                 `json:"pipeline_id"`
	Payload    map[string]interface{} `json:"payload"`
	Timestamp  time.Time              `json:"timestamp"`
}

type EventStore struct {
	mu     sync.RWMutex
	events []Event
	nextID int
}

func NewEventStore() *EventStore {
	return &EventStore{events: make([]Event, 0)}
}

func (s *EventStore) Add(pipelineID string, payload map[string]interface{}) Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	evt := Event{
		ID:         fmt.Sprintf("evt-%d", s.nextID),
		PipelineID: pipelineID,
		Payload:    payload,
		Timestamp:  time.Now(),
	}
	s.events = append(s.events, evt)
	return evt
}

func (s *EventStore) List(pipelineID string) []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if pipelineID == "" {
		result := make([]Event, len(s.events))
		copy(result, s.events)
		return result
	}
	var result []Event
	for _, e := range s.events {
		if e.PipelineID == pipelineID {
			result = append(result, e)
		}
	}
	return result
}

func (s *EventStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.events)
}

var store = NewEventStore()

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "ingestion-gateway"})
}

func ingestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var body struct {
		PipelineID string                 `json:"pipeline_id"`
		Payload    map[string]interface{} `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if body.PipelineID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "pipeline_id is required"})
		return
	}
	evt := store.Add(body.PipelineID, body.Payload)
	log.Printf("Ingested event %s for pipeline %s", evt.ID, evt.PipelineID)
	writeJSON(w, http.StatusCreated, evt)
}

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	pipelineID := r.URL.Query().Get("pipeline_id")
	events := store.List(pipelineID)
	log.Printf("Listed %d events (filter: pipeline_id=%q)", len(events), pipelineID)
	writeJSON(w, http.StatusOK, events)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_events": store.Count(),
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func main() {
	port := os.Getenv("INGESTION_PORT")
	if port == "" {
		port = "8002"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/ingest", ingestHandler)
	mux.HandleFunc("/api/events", eventsHandler)
	mux.HandleFunc("/api/stats", statsHandler)

	log.Printf("Starting ingestion-gateway on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
