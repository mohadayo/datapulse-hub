package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func resetStore() {
	store = NewEventStore()
}

func TestHealthHandler(t *testing.T) {
	resetStore()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Fatalf("expected status ok, got %s", body["status"])
	}
	if body["service"] != "ingestion-gateway" {
		t.Fatalf("expected service ingestion-gateway, got %s", body["service"])
	}
}

func TestIngestHandler(t *testing.T) {
	resetStore()
	payload := `{"pipeline_id":"p1","payload":{"key":"value"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/ingest", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ingestHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	var evt Event
	json.NewDecoder(w.Body).Decode(&evt)
	if evt.PipelineID != "p1" {
		t.Fatalf("expected pipeline_id p1, got %s", evt.PipelineID)
	}
	if evt.ID != "evt-1" {
		t.Fatalf("expected id evt-1, got %s", evt.ID)
	}
}

func TestIngestHandlerMissingPipelineID(t *testing.T) {
	resetStore()
	payload := `{"payload":{"key":"value"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/ingest", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ingestHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestIngestHandlerInvalidJSON(t *testing.T) {
	resetStore()
	req := httptest.NewRequest(http.MethodPost, "/api/ingest", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ingestHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestIngestHandlerMethodNotAllowed(t *testing.T) {
	resetStore()
	req := httptest.NewRequest(http.MethodGet, "/api/ingest", nil)
	w := httptest.NewRecorder()
	ingestHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestEventsHandler(t *testing.T) {
	resetStore()
	store.Add("p1", map[string]interface{}{"a": 1})
	store.Add("p2", map[string]interface{}{"b": 2})
	store.Add("p1", map[string]interface{}{"c": 3})

	req := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	w := httptest.NewRecorder()
	eventsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var events []Event
	json.NewDecoder(w.Body).Decode(&events)
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
}

func TestEventsHandlerFilterByPipeline(t *testing.T) {
	resetStore()
	store.Add("p1", map[string]interface{}{"a": 1})
	store.Add("p2", map[string]interface{}{"b": 2})

	req := httptest.NewRequest(http.MethodGet, "/api/events?pipeline_id=p1", nil)
	w := httptest.NewRecorder()
	eventsHandler(w, req)

	var events []Event
	json.NewDecoder(w.Body).Decode(&events)
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].PipelineID != "p1" {
		t.Fatalf("expected pipeline_id p1, got %s", events[0].PipelineID)
	}
}

func TestStatsHandler(t *testing.T) {
	resetStore()
	store.Add("p1", map[string]interface{}{})
	store.Add("p1", map[string]interface{}{})

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w := httptest.NewRecorder()
	statsHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var body map[string]interface{}
	json.NewDecoder(w.Body).Decode(&body)
	if body["total_events"].(float64) != 2 {
		t.Fatalf("expected 2 total events, got %v", body["total_events"])
	}
}
