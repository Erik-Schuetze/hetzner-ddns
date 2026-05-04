package hetzner

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type RRSetRecord struct {
	Value string `json:"value"`
}

type RRSet struct {
	Name    string
	Type    string
	TTL     int
	Records []RRSetRecord
}

type rrsetResponse struct {
	RRSet struct {
		Name    string        `json:"name"`
		Type    string        `json:"type"`
		TTL     int           `json:"ttl"`
		Records []RRSetRecord `json:"records"`
	} `json:"rrset"`
}

type apiErrorResponse struct {
	Error struct {
		Code    any    `json:"code"`
		Message string `json:"message"`
		Details any    `json:"details,omitempty"`
	} `json:"error"`
}

type actionResponse struct {
	Action struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
		Error  any    `json:"error,omitempty"`
	} `json:"action"`
}

var (
	BaseURL          = "https://api.hetzner.cloud/v1"
	ErrRRSetNotFound = errors.New("rrset not found")
)

const (
	actionPollInterval = 500 * time.Millisecond
	maxActionPolls     = 20
)

func GetAPIToken() string {
	apiToken := os.Getenv("HETZNER_CLOUD_API_TOKEN")
	if apiToken == "" {
		panic("Missing HETZNER_CLOUD_API_TOKEN env variable")
	}
	return apiToken
}

func GetRRSet(zoneName, recordName, recordType string) (RRSet, error) {
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, buildRRSetURL(zoneName, recordName, recordType), nil)
	if err != nil {
		return RRSet{}, fmt.Errorf("failed to create request: %w", err)
	}
	setHeaders(req, false)

	resp, err := client.Do(req)
	if err != nil {
		return RRSet{}, fmt.Errorf("request failed: %w", err)
	}
	defer closeResponseBody(resp)

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return RRSet{}, ErrRRSetNotFound
	default:
		return RRSet{}, parseAPIError("failed to get rrset", resp)
	}

	var response rrsetResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return RRSet{}, fmt.Errorf("failed to decode response: %w", err)
	}

	return RRSet{
		Name:    response.RRSet.Name,
		Type:    response.RRSet.Type,
		TTL:     response.RRSet.TTL,
		Records: response.RRSet.Records,
	}, nil
}

func CreateRRSet(zoneName, recordName, recordType string, ttl int, values []string) error {
	payload := struct {
		Name    string        `json:"name"`
		Type    string        `json:"type"`
		TTL     int           `json:"ttl"`
		Records []RRSetRecord `json:"records"`
	}{
		Name:    recordName,
		Type:    strings.ToUpper(recordType),
		TTL:     ttl,
		Records: buildRRSetRecords(values),
	}

	return doActionRequest(
		http.MethodPost,
		buildRRSetCollectionURL(zoneName),
		payload,
		http.StatusCreated,
		http.StatusAccepted,
	)
}

func SetRRSetRecords(zoneName, recordName, recordType string, values []string) error {
	payload := struct {
		Records []RRSetRecord `json:"records"`
	}{
		Records: buildRRSetRecords(values),
	}

	return doActionRequest(
		http.MethodPost,
		buildRRSetActionURL(zoneName, recordName, recordType, "set_records"),
		payload,
		http.StatusCreated,
		http.StatusAccepted,
	)
}

func ChangeRRSetTTL(zoneName, recordName, recordType string, ttl int) error {
	payload := struct {
		TTL int `json:"ttl"`
	}{
		TTL: ttl,
	}

	return doActionRequest(
		http.MethodPost,
		buildRRSetActionURL(zoneName, recordName, recordType, "change_ttl"),
		payload,
		http.StatusAccepted,
	)
}

func doActionRequest(method, endpoint string, payload any, successCodes ...int) error {
	var body io.Reader
	if payload != nil {
		buffer := bytes.NewBuffer(nil)
		if err := json.NewEncoder(buffer).Encode(payload); err != nil {
			return fmt.Errorf("failed to encode request payload: %w", err)
		}
		body = buffer
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	setHeaders(req, payload != nil)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer closeResponseBody(resp)

	if !containsStatus(successCodes, resp.StatusCode) {
		return parseAPIError("request failed", resp)
	}

	var actionResp actionResponse
	if err := json.NewDecoder(resp.Body).Decode(&actionResp); err != nil {
		return fmt.Errorf("failed to decode action response: %w", err)
	}

	return waitForAction(client, actionResp)
}

func setHeaders(req *http.Request, hasBody bool) {
	req.Header.Set("Authorization", "Bearer "+GetAPIToken())
	req.Header.Set("Accept", "application/json")
	if hasBody {
		req.Header.Set("Content-Type", "application/json")
	}
}

func buildRRSetURL(zoneName, recordName, recordType string) string {
	return fmt.Sprintf(
		"%s/zones/%s/rrsets/%s/%s",
		BaseURL,
		url.PathEscape(zoneName),
		url.PathEscape(recordName),
		url.PathEscape(strings.ToUpper(recordType)),
	)
}

func buildRRSetCollectionURL(zoneName string) string {
	return fmt.Sprintf("%s/zones/%s/rrsets", BaseURL, url.PathEscape(zoneName))
}

func buildRRSetActionURL(zoneName, recordName, recordType, action string) string {
	return fmt.Sprintf("%s/actions/%s", buildRRSetURL(zoneName, recordName, recordType), action)
}

func buildRRSetRecords(values []string) []RRSetRecord {
	records := make([]RRSetRecord, 0, len(values))
	for _, value := range values {
		records = append(records, RRSetRecord{Value: value})
	}
	return records
}

func containsStatus(expected []int, actual int) bool {
	for _, code := range expected {
		if code == actual {
			return true
		}
	}
	return false
}

func waitForAction(client *http.Client, actionResp actionResponse) error {
	switch actionResp.Action.Status {
	case "success":
		return nil
	case "running":
	case "queued":
	default:
		if actionResp.Action.Error != nil {
			return fmt.Errorf("action %d failed: %v", actionResp.Action.ID, actionResp.Action.Error)
		}
		return fmt.Errorf("unexpected action status %q", actionResp.Action.Status)
	}

	if actionResp.Action.ID == 0 {
		return errors.New("missing action id in response")
	}

	for range maxActionPolls {
		time.Sleep(actionPollInterval)

		req, err := http.NewRequest(http.MethodGet, buildActionURL(actionResp.Action.ID), nil)
		if err != nil {
			return fmt.Errorf("failed to create action poll request: %w", err)
		}
		setHeaders(req, false)

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("action poll request failed: %w", err)
		}

		var pollResp actionResponse
		if resp.StatusCode == http.StatusOK {
			err = json.NewDecoder(resp.Body).Decode(&pollResp)
			closeResponseBody(resp)
			if err != nil {
				return fmt.Errorf("failed to decode action poll response: %w", err)
			}

			switch pollResp.Action.Status {
			case "success":
				return nil
			case "running", "queued":
				continue
			default:
				if pollResp.Action.Error != nil {
					return fmt.Errorf("action %d failed: %v", pollResp.Action.ID, pollResp.Action.Error)
				}
				return fmt.Errorf("unexpected action status %q", pollResp.Action.Status)
			}
		}

		err = parseAPIError("action poll failed", resp)
		closeResponseBody(resp)
		return err
	}

	return fmt.Errorf("action %d did not complete after %d polls", actionResp.Action.ID, maxActionPolls)
}

func buildActionURL(actionID int64) string {
	return fmt.Sprintf("%s/zones/actions/%d", BaseURL, actionID)
}

func parseAPIError(action string, resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s: %s", action, resp.Status)
	}

	var apiErr apiErrorResponse
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error.Message != "" {
		if apiErr.Error.Details != nil {
			return fmt.Errorf("%s: %s: %s (%v)", action, resp.Status, apiErr.Error.Message, apiErr.Error.Details)
		}
		return fmt.Errorf("%s: %s: %s", action, resp.Status, apiErr.Error.Message)
	}

	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return fmt.Errorf("%s: %s", action, resp.Status)
	}

	return fmt.Errorf("%s: %s: %s", action, resp.Status, trimmed)
}

func closeResponseBody(resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.Printf("Error closing response body: %v", err)
	}
}
