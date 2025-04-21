package ipdetect

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPublicIP(t *testing.T) {
	tests := []struct {
		name      string
		responses map[string]struct {
			status int
			body   string
			err    error
		}
		wantIP  string
		wantErr bool
	}{
		{
			name: "successful first service",
			responses: map[string]struct {
				status int
				body   string
				err    error
			}{
				"https://checkip.amazonaws.com": {status: 200, body: "1.2.3.4\n", err: nil},
			},
			wantIP:  "1.2.3.4",
			wantErr: false,
		},
		{
			name: "first service fails, second succeeds",
			responses: map[string]struct {
				status int
				body   string
				err    error
			}{
				"https://checkip.amazonaws.com": {status: 500, body: "", err: nil},
				"https://api.ipify.org":         {status: 200, body: "5.6.7.8", err: nil},
			},
			wantIP:  "5.6.7.8",
			wantErr: false,
		},
		{
			name: "all services fail",
			responses: map[string]struct {
				status int
				body   string
				err    error
			}{
				"https://checkip.amazonaws.com": {status: 500, body: "", err: nil},
				"https://api.ipify.org":         {status: 500, body: "", err: nil},
				"https://icanhazip.com":         {status: 500, body: "", err: nil},
			},
			wantIP:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original services
			originalServices := ipCheckServices
			defer func() { ipCheckServices = originalServices }()

			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Get the response for the current service being tested
				var currentResponse struct {
					status int
					body   string
					err    error
				}

				// Find the first matching response
				for _, resp := range tt.responses {
					currentResponse = resp
					break
				}

				if currentResponse.err != nil {
					http.Error(w, currentResponse.err.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(currentResponse.status)
				fmt.Fprint(w, currentResponse.body)
			}))
			defer server.Close()

			// Replace real services with test server
			ipCheckServices = []string{server.URL}

			// Run the test
			gotIP, err := GetPublicIP()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPublicIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotIP != tt.wantIP {
				t.Errorf("GetPublicIP() = %v, want %v", gotIP, tt.wantIP)
			}
		})
	}
}

func TestFetchIP(t *testing.T) {
	tests := []struct {
		name     string
		response string
		status   int
		wantIP   string
		wantErr  bool
	}{
		{
			name:     "valid IP",
			response: "1.2.3.4\n",
			status:   http.StatusOK,
			wantIP:   "1.2.3.4",
			wantErr:  false,
		},
		{
			name:     "server error",
			response: "Internal Server Error",
			status:   http.StatusInternalServerError,
			wantIP:   "",
			wantErr:  true,
		},
		{
			name:     "invalid IP format",
			response: "not an ip",
			status:   http.StatusOK,
			wantIP:   "not an ip",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.status != http.StatusOK {
					http.Error(w, tt.response, tt.status)
					return
				}
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.response)
			}))
			defer server.Close()

			gotIP, err := fetchIP(server.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotIP != tt.wantIP {
				t.Errorf("fetchIP() = %v, want %v", gotIP, tt.wantIP)
			}
		})
	}
}
