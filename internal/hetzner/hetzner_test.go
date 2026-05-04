package hetzner_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erik-schuetze/hetzner-ddns/internal/hetzner"
)

func TestGetRRSet(t *testing.T) {
	originalBaseURL := hetzner.BaseURL
	defer func() { hetzner.BaseURL = originalBaseURL }()

	t.Setenv("HETZNER_CLOUD_API_TOKEN", "test-token")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("Authorization header = %q, want %q", got, "Bearer test-token")
		}

		switch r.URL.Path {
		case "/zones/example.com/rrsets/@/A":
			if r.Method != http.MethodGet {
				t.Fatalf("method = %s, want GET", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]any{
				"rrset": map[string]any{
					"name": "@",
					"type": "A",
					"ttl":  3600,
					"records": []map[string]any{
						{"value": "1.2.3.4"},
					},
				},
			}); err != nil {
				t.Fatalf("encoding response: %v", err)
			}
		case "/zones/example.com/rrsets/missing/A":
			http.NotFound(w, r)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	hetzner.BaseURL = server.URL

	rrset, err := hetzner.GetRRSet("example.com", "@", "A")
	if err != nil {
		t.Fatalf("GetRRSet() error = %v", err)
	}

	if rrset.Name != "@" {
		t.Fatalf("rrset.Name = %q, want %q", rrset.Name, "@")
	}
	if rrset.Type != "A" {
		t.Fatalf("rrset.Type = %q, want %q", rrset.Type, "A")
	}
	if rrset.TTL != 3600 {
		t.Fatalf("rrset.TTL = %d, want %d", rrset.TTL, 3600)
	}
	if len(rrset.Records) != 1 || rrset.Records[0].Value != "1.2.3.4" {
		t.Fatalf("rrset.Records = %#v, want single value 1.2.3.4", rrset.Records)
	}

	_, err = hetzner.GetRRSet("example.com", "missing", "A")
	if !errors.Is(err, hetzner.ErrRRSetNotFound) {
		t.Fatalf("GetRRSet() missing error = %v, want ErrRRSetNotFound", err)
	}
}

func TestSetRRSetRecords(t *testing.T) {
	originalBaseURL := hetzner.BaseURL
	defer func() { hetzner.BaseURL = originalBaseURL }()

	t.Setenv("HETZNER_CLOUD_API_TOKEN", "test-token")

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		switch requestCount {
		case 1:
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			if got := r.URL.EscapedPath(); got != "/zones/example.com/rrsets/www/A/actions/set_records" {
				t.Fatalf("path = %s, want %s", got, "/zones/example.com/rrsets/www/A/actions/set_records")
			}
			if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
				t.Fatalf("Authorization header = %q, want %q", got, "Bearer test-token")
			}
			if got := r.Header.Get("Content-Type"); got != "application/json" {
				t.Fatalf("Content-Type = %q, want %q", got, "application/json")
			}

			var payload struct {
				Records []hetzner.RRSetRecord `json:"records"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			if len(payload.Records) != 1 || payload.Records[0].Value != "1.2.3.4" {
				t.Fatalf("payload.Records = %#v, want single value 1.2.3.4", payload.Records)
			}

			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(map[string]any{
				"action": map[string]any{
					"id":     42,
					"status": "running",
				},
			}); err != nil {
				t.Fatalf("encoding action response: %v", err)
			}
		case 2:
			if r.Method != http.MethodGet {
				t.Fatalf("poll method = %s, want GET", r.Method)
			}
			if got := r.URL.EscapedPath(); got != "/zones/actions/42" {
				t.Fatalf("poll path = %s, want %s", got, "/zones/actions/42")
			}
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]any{
				"action": map[string]any{
					"id":     42,
					"status": "success",
				},
			}); err != nil {
				t.Fatalf("encoding action poll response: %v", err)
			}
		default:
			t.Fatalf("unexpected request count: %d", requestCount)
		}
	}))
	defer server.Close()

	hetzner.BaseURL = server.URL

	if err := hetzner.SetRRSetRecords("example.com", "www", "A", []string{"1.2.3.4"}); err != nil {
		t.Fatalf("SetRRSetRecords() error = %v", err)
	}
}

func TestCreateRRSet(t *testing.T) {
	originalBaseURL := hetzner.BaseURL
	defer func() { hetzner.BaseURL = originalBaseURL }()

	t.Setenv("HETZNER_CLOUD_API_TOKEN", "test-token")

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		switch requestCount {
		case 1:
			if r.Method != http.MethodPost {
				t.Fatalf("method = %s, want POST", r.Method)
			}
			if got := r.URL.EscapedPath(); got != "/zones/example.com/rrsets" {
				t.Fatalf("path = %s, want %s", got, "/zones/example.com/rrsets")
			}

			var payload struct {
				Name    string                `json:"name"`
				Type    string                `json:"type"`
				TTL     int                   `json:"ttl"`
				Records []hetzner.RRSetRecord `json:"records"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			if payload.Name != "mail" || payload.Type != "A" || payload.TTL != 4000 {
				t.Fatalf("payload = %#v, want name mail, type A, ttl 4000", payload)
			}
			if len(payload.Records) != 1 || payload.Records[0].Value != "1.2.3.4" {
				t.Fatalf("payload.Records = %#v, want single value 1.2.3.4", payload.Records)
			}

			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(map[string]any{
				"action": map[string]any{
					"id":     77,
					"status": "running",
				},
			}); err != nil {
				t.Fatalf("encoding action response: %v", err)
			}
		case 2:
			if r.Method != http.MethodGet {
				t.Fatalf("poll method = %s, want GET", r.Method)
			}
			if got := r.URL.EscapedPath(); got != "/zones/actions/77" {
				t.Fatalf("poll path = %s, want %s", got, "/zones/actions/77")
			}
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]any{
				"action": map[string]any{
					"id":     77,
					"status": "success",
				},
			}); err != nil {
				t.Fatalf("encoding action poll response: %v", err)
			}
		default:
			t.Fatalf("unexpected request count: %d", requestCount)
		}
	}))
	defer server.Close()

	hetzner.BaseURL = server.URL

	if err := hetzner.CreateRRSet("example.com", "mail", "A", 4000, []string{"1.2.3.4"}); err != nil {
		t.Fatalf("CreateRRSet() error = %v", err)
	}
}

func TestChangeRRSetTTL(t *testing.T) {
	originalBaseURL := hetzner.BaseURL
	defer func() { hetzner.BaseURL = originalBaseURL }()

	t.Setenv("HETZNER_CLOUD_API_TOKEN", "test-token")

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		switch requestCount {
		case 1:
			if r.Method != http.MethodPost {
				t.Fatalf("first method = %s, want POST", r.Method)
			}
			if got := r.URL.EscapedPath(); got != "/zones/example.com/rrsets/www/A/actions/change_ttl" {
				t.Fatalf("first path = %s, want %s", got, "/zones/example.com/rrsets/www/A/actions/change_ttl")
			}

			var payload struct {
				TTL int `json:"ttl"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			if payload.TTL != 120 {
				t.Fatalf("payload.TTL = %d, want %d", payload.TTL, 120)
			}

			w.WriteHeader(http.StatusAccepted)
			if err := json.NewEncoder(w).Encode(map[string]any{
				"action": map[string]any{
					"id":     99,
					"status": "running",
				},
			}); err != nil {
				t.Fatalf("encoding action response: %v", err)
			}
		case 2:
			if r.Method != http.MethodGet {
				t.Fatalf("second method = %s, want GET", r.Method)
			}
			if got := r.URL.EscapedPath(); got != "/zones/actions/99" {
				t.Fatalf("second path = %s, want %s", got, "/zones/actions/99")
			}
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]any{
				"action": map[string]any{
					"id":     99,
					"status": "success",
				},
			}); err != nil {
				t.Fatalf("encoding action poll response: %v", err)
			}
		default:
			t.Fatalf("unexpected request count: %d", requestCount)
		}
	}))
	defer server.Close()

	hetzner.BaseURL = server.URL

	if err := hetzner.ChangeRRSetTTL("example.com", "www", "A", 120); err != nil {
		t.Fatalf("ChangeRRSetTTL() error = %v", err)
	}
}

func TestGetAPIToken(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		t.Setenv("HETZNER_CLOUD_API_TOKEN", "test-token")

		got := hetzner.GetAPIToken()
		if got != "test-token" {
			t.Fatalf("GetAPIToken() = %q, want %q", got, "test-token")
		}
	})

	t.Run("missing token", func(t *testing.T) {
		t.Setenv("HETZNER_CLOUD_API_TOKEN", "")

		defer func() {
			if r := recover(); r == nil {
				t.Fatal("GetAPIToken() expected panic for missing token")
			}
		}()

		hetzner.GetAPIToken()
	})
}
