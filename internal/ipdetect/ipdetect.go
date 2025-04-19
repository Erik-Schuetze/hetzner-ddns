package ipdetect

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// List of services to try in order
var ipCheckServices = []string{
	"https://checkip.amazonaws.com",
	"https://api.ipify.org",
	"https://icanhazip.com",
}

func GetPublicIP() (string, error) {
	var lastErr error

	for _, service := range ipCheckServices {
		ip, err := fetchIP(service)
		if err != nil {
			lastErr = err
			continue
		}
		return ip, nil
	}

	return "", fmt.Errorf("all IP detection services failed, last error: %v", lastErr)
}

func fetchIP(service string) (string, error) {
	resp, err := http.Get(service)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Trim whitespace and newlines
	return strings.TrimSpace(string(body)), nil
}
