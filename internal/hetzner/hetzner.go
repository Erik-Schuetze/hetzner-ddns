package hetzner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func GetApiToken() string {
	apiToken := os.Getenv("HETZNER_API_TOKEN")
	if apiToken == "" {
		panic("Missing HETZNER_API_TOKEN env variable")
	}
	return apiToken
}

func GetRecordById(recordID string) {
	// Get Record (GET https://dns.hetzner.com/api/v1/records/{RecordID})

	// Create client
	client := &http.Client{}

	// Create request
	url := fmt.Sprintf("https://dns.hetzner.com/api/v1/records/%s", recordID)
	req, err := http.NewRequest("GET", url, nil)

	// Headers
	apiToken := GetApiToken()
	req.Header.Add("Auth-API-Token", apiToken)

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}

	// Read Response Body
	respBody, _ := io.ReadAll(resp.Body)

	// Display Results
	fmt.Println("response Status : ", resp.Status)
	fmt.Println("response Headers : ", resp.Header)
	fmt.Println("response Body : ", string(respBody))
}

func GetAllRecordsByZone(zoneID string) ([]Record, error) {
	// Get Records (GET https://dns.hetzner.com/api/v1/records?zone_id={ZoneID})

	// Create client
	client := &http.Client{}

	// Create request
	url := fmt.Sprintf("https://dns.hetzner.com/api/v1/records?zone_id=%s", zoneID)
	req, err := http.NewRequest("GET", url, nil)

	// Headers
	apiToken := GetApiToken()
	req.Header.Add("Auth-API-Token", apiToken)

	parseFormErr := req.ParseForm()
	if parseFormErr != nil {
		fmt.Println(parseFormErr)
	}

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	var response RecordsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return response.Records, nil
}

func UpdateRecord(record Record) error {
	// Update Record (PUT https://dns.hetzner.com/api/v1/records/{RecordID})
	jsonContent := fmt.Sprintf(`{"value": "%s","ttl": %d,"type": "%s","name": "%s","zone_id": "%s"}`, record.Value, record.TTL, record.Type, record.Name, record.ZoneID)
	json := []byte(jsonContent)
	body := bytes.NewBuffer(json)

	// Create client
	client := &http.Client{}

	// Create request
	url := fmt.Sprintf("https://dns.hetzner.com/api/v1/records/%s", record.ID)
	req, err := http.NewRequest("PUT", url, body)

	// Headers
	req.Header.Add("Content-Type", "application/json")
	apiToken := GetApiToken()
	req.Header.Add("Auth-API-Token", apiToken)

	// Fetch Request
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("Failure : ", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update record: %s", resp.Status)
	} else {
		return nil
	}
}
