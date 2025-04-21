package hetzner_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/erik-schuetze/hetzner-ddns/internal/hetzner"
)

func TestGetAllRecordsByZone(t *testing.T) {
	// Setup test server
	records := []hetzner.Record{
		{
			ID:      "1",
			Type:    "A",
			Name:    "test",
			Value:   "1.2.3.4",
			ZoneID:  "zone1",
			TTL:     3600,
			Created: "2023-01-01",
		},
	}

	// Save original API URL and restore after test
	originalBaseURL := hetzner.BaseURL
	defer func() { hetzner.BaseURL = originalBaseURL }()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.Header.Get("Auth-API-Token") == "" {
			t.Error("Missing Auth-API-Token header")
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(hetzner.RecordsResponse{Records: records})
	}))
	defer server.Close()

	// Set test server URL as base URL
	hetzner.BaseURL = server.URL

	// Set test environment
	os.Setenv("HETZNER_API_TOKEN", "test-token")

	// Test successful case
	t.Run("successful records retrieval", func(t *testing.T) {
		got, err := hetzner.GetAllRecordsByZone("zone1")
		if err != nil {
			t.Fatalf("GetAllRecordsByZone() error = %v", err)
		}
		if len(got) != len(records) {
			t.Errorf("GetAllRecordsByZone() got %v records, want %v", len(got), len(records))
		}
	})
}

func TestUpdateRecord(t *testing.T) {
	// Save original API URL and restore after test
	originalBaseURL := hetzner.BaseURL
	defer func() { hetzner.BaseURL = originalBaseURL }()

	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		if r.Header.Get("Auth-API-Token") == "" {
			t.Error("Missing Auth-API-Token header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Missing Content-Type header")
		}

		// Return success
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Set test server URL as base URL
	hetzner.BaseURL = server.URL

	// Set test environment
	os.Setenv("HETZNER_API_TOKEN", "test-token")

	// Test successful update
	t.Run("successful record update", func(t *testing.T) {
		record := hetzner.Record{
			ID:     "test-id",
			Type:   "A",
			Name:   "test.example.com",
			Value:  "1.2.3.4",
			ZoneID: "zone1",
			TTL:    3600,
		}

		err := hetzner.UpdateRecord(record)
		if err != nil {
			t.Fatalf("UpdateRecord() error = %v", err)
		}
	})

	// Test error cases
	t.Run("invalid record", func(t *testing.T) {
		record := hetzner.Record{} // Empty record
		err := hetzner.UpdateRecord(record)
		if err == nil {
			t.Error("UpdateRecord() expected error for invalid record")
		}
	})

	// Test error cases
	t.Run("invalid records", func(t *testing.T) {
		tests := []struct {
			name   string
			record hetzner.Record
		}{
			{
				name:   "empty record",
				record: hetzner.Record{},
			},
			{
				name: "missing ID",
				record: hetzner.Record{
					Type:   "A",
					Name:   "test",
					ZoneID: "zone1",
				},
			},
			{
				name: "missing Type",
				record: hetzner.Record{
					ID:     "1",
					Name:   "test",
					ZoneID: "zone1",
				},
			},
			{
				name: "missing Name",
				record: hetzner.Record{
					ID:     "1",
					Type:   "A",
					ZoneID: "zone1",
				},
			},
			{
				name: "missing ZoneID",
				record: hetzner.Record{
					ID:   "1",
					Type: "A",
					Name: "test",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if err := hetzner.UpdateRecord(tt.record); err == nil {
					t.Errorf("UpdateRecord() expected error for %s", tt.name)
				}
			})
		}
	})
}

func TestGetApiToken(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		expected := "test-token"
		os.Setenv("HETZNER_API_TOKEN", expected)

		got := hetzner.GetApiToken()
		if got != expected {
			t.Errorf("GetApiToken() = %v, want %v", got, expected)
		}
	})

	t.Run("missing token", func(t *testing.T) {
		os.Unsetenv("HETZNER_API_TOKEN")

		defer func() {
			if r := recover(); r == nil {
				t.Error("GetApiToken() expected panic for missing token")
			}
		}()

		hetzner.GetApiToken()
	})
}
