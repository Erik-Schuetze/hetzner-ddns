package controller

import (
	"errors"
	"testing"

	"github.com/erik-schuetze/hetzner-ddns/internal/config"
	"github.com/erik-schuetze/hetzner-ddns/internal/hetzner"
)

func TestReconcileCreatesMissingConfiguredRecord(t *testing.T) {
	originalDetectPublicIP := detectPublicIP
	originalGetRRSet := getRRSet
	originalCreateRRSet := createRRSet
	t.Cleanup(func() {
		detectPublicIP = originalDetectPublicIP
		getRRSet = originalGetRRSet
		createRRSet = originalCreateRRSet
	})

	detectPublicIP = func() (string, error) {
		return "1.2.3.4", nil
	}

	cfg := &config.Config{}
	cfg.Params.RefreshInterval = 15
	cfg.Hetzner.Zones = []config.Zone{
		{
			Name: "example.com",
			Records: []config.Record{
				{Name: "mail", Type: "A", TTL: 4000},
			},
		},
	}

	getRRSet = func(zoneName, recordName, recordType string) (hetzner.RRSet, error) {
		if zoneName != "example.com" || recordName != "mail" || recordType != "A" {
			t.Fatalf("unexpected getRRSet input: %s %s %s", zoneName, recordName, recordType)
		}
		return hetzner.RRSet{}, hetzner.ErrRRSetNotFound
	}

	createCalled := false
	createRRSet = func(zoneName, recordName, recordType string, ttl int, values []string) error {
		createCalled = true
		if zoneName != "example.com" || recordName != "mail" || recordType != "A" {
			t.Fatalf("unexpected createRRSet input: %s %s %s", zoneName, recordName, recordType)
		}
		if ttl != 4000 {
			t.Fatalf("ttl = %d, want 4000", ttl)
		}
		if len(values) != 1 || values[0] != "1.2.3.4" {
			t.Fatalf("values = %#v, want single value 1.2.3.4", values)
		}
		return nil
	}

	c := NewController(cfg, "/tmp/config.yaml")
	if err := c.reconcile(); err != nil {
		t.Fatalf("reconcile() error = %v", err)
	}

	if !createCalled {
		t.Fatal("createRRSet was not called")
	}
}

func TestReconcileContinuesOnCreateError(t *testing.T) {
	originalDetectPublicIP := detectPublicIP
	originalGetRRSet := getRRSet
	originalCreateRRSet := createRRSet
	t.Cleanup(func() {
		detectPublicIP = originalDetectPublicIP
		getRRSet = originalGetRRSet
		createRRSet = originalCreateRRSet
	})

	detectPublicIP = func() (string, error) {
		return "1.2.3.4", nil
	}

	cfg := &config.Config{}
	cfg.Hetzner.Zones = []config.Zone{
		{
			Name: "example.com",
			Records: []config.Record{
				{Name: "mail", Type: "A", TTL: 4000},
			},
		},
	}

	getRRSet = func(zoneName, recordName, recordType string) (hetzner.RRSet, error) {
		return hetzner.RRSet{}, hetzner.ErrRRSetNotFound
	}

	createRRSet = func(zoneName, recordName, recordType string, ttl int, values []string) error {
		return errors.New("boom")
	}

	c := NewController(cfg, "/tmp/config.yaml")
	if err := c.reconcile(); err != nil {
		t.Fatalf("reconcile() error = %v, want nil because create failures are logged and skipped", err)
	}
}
