package hetzner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Record struct {
	ID       string `json:"id" yaml:"-"`
	Type     string `json:"type" yaml:"type"`
	Name     string `json:"name" yaml:"name"`
	Value    string `json:"value" yaml:"value,omitempty"`
	ZoneID   string `json:"zone_id" yaml:"-"`
	TTL      int    `json:"ttl,omitempty" yaml:"ttl"`
	Created  string `json:"created" yaml:"-"`
	Modified string `json:"modified" yaml:"-"`
}

type RecordsResponse struct {
	Records []Record `json:"records"`
}

var BaseURL = "https://dns.hetzner.com/api/v1"

func GetApiToken() string {
	apiToken := os.Getenv("HETZNER_API_TOKEN")
	if apiToken == "" {
		panic("Missing HETZNER_API_TOKEN env variable")
	}
	return apiToken
}

func GetAllRecordsByZone(zoneID string) ([]Record, error) {
	client := &http.Client{}
	url := fmt.Sprintf("%s/records?zone_id=%s", BaseURL, zoneID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Auth-API-Token", GetApiToken())

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	var response RecordsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Records, nil
}

func UpdateRecord(record Record) error {
	// Validate required fields
	if record.ID == "" || record.Type == "" || record.Name == "" || record.ZoneID == "" {
		return fmt.Errorf("invalid record: missing required fields")
	}

	jsonContent := fmt.Sprintf(`{"value":"%s","ttl":%d,"type":"%s","name":"%s","zone_id":"%s"}`,
		record.Value, record.TTL, record.Type, record.Name, record.ZoneID)
	body := bytes.NewBufferString(jsonContent)

	client := &http.Client{}
	url := fmt.Sprintf("%s/records/%s", BaseURL, record.ID)

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Auth-API-Token", GetApiToken())

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update record: %s", resp.Status)
	}
	return nil
}
