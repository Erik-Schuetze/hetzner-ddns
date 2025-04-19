package controller

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/erik-schuetze/hetzner-ddns/internal/config"
	"github.com/erik-schuetze/hetzner-ddns/internal/hetzner"
	"github.com/erik-schuetze/hetzner-ddns/internal/ipdetect"
)

type Controller struct {
	config     *config.Config
	configPath string
	stopCh     chan struct{}
	mu         sync.RWMutex // Protects config
}

func NewController(cfg *config.Config, configPath string) *Controller {
	return &Controller{
		config:     cfg,
		configPath: configPath,
		stopCh:     make(chan struct{}),
	}
}

// Run starts the reconciliation loop and handles graceful shutdown
func (c *Controller) Run(ctx context.Context) error {
	// Start a goroutine to query the API and handle ddns updates
	ticker := time.NewTicker(time.Duration(c.config.Params.RefreshInterval) * time.Minute)
	defer ticker.Stop()

	// Do initial reconciliation
	if err := c.reconcile(); err != nil {
		log.Printf("Error in initial reconciliation: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down gracefully...")
			return nil
		case <-ticker.C:
			// Reload config first
			if err := c.reloadConfig(); err != nil {
				log.Printf("Failed to reload config: %v", err)
				continue
			}

			// Then reconcile
			if err := c.reconcile(); err != nil {
				log.Printf("Error in reconciliation: %v", err)
			}
		}
	}
}

func (c *Controller) reloadConfig() error {
	// Load the new config
	newConfig, err := config.Load(c.configPath)
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	// lock the config for writing
	c.mu.Lock()
	c.config = newConfig
	c.mu.Unlock()

	return nil
}

func (c *Controller) reconcile() error {
	// Get the public IP address
	ip, err := ipdetect.GetPublicIP()
	if err != nil {
		return fmt.Errorf("failed to detect public IP: %w", err)
	}

	// lock the config for reading
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, zone := range c.config.Hetzner.Zones {
		hetznerRecords, err := hetzner.GetAllRecordsByZone(zone.ZoneID)
		if err != nil {
			return fmt.Errorf("failed to get records for zone %s: %w", zone.ZoneID, err)
		}

		for _, configRecord := range zone.Records {
			for _, hetznerRecord := range hetznerRecords {
				if configRecord.Name == hetznerRecord.Name && configRecord.Type == hetznerRecord.Type {
					if err := c.updateRecordIfNeeded(zone.ZoneID, configRecord, hetznerRecord, ip); err != nil {
						log.Printf("Error updating record %s: %v", configRecord.Name, err)
						continue
					}
				}
			}
		}
	}

	return nil
}

// Helper function to make the code more readable
func (c *Controller) updateRecordIfNeeded(zoneID string, configRecord hetzner.Record, hetznerRecord hetzner.Record, ip string) error {
	if configRecord.TTL != hetznerRecord.TTL {
		hetznerRecord.TTL = configRecord.TTL
		if err := hetzner.UpdateRecord(hetznerRecord); err != nil {
			return fmt.Errorf("failed to update TTL: %w", err)
		}
		log.Printf("Updated TTL - zone: %s, name: %s, TTL: %d", zoneID, configRecord.Name, hetznerRecord.TTL)
	}

	if ip != hetznerRecord.Value {
		hetznerRecord.Value = ip
		if err := hetzner.UpdateRecord(hetznerRecord); err != nil {
			return fmt.Errorf("failed to update IP: %w", err)
		}
		log.Printf("Updated Value - zone: %s, name: %s, value: %s", zoneID, configRecord.Name, hetznerRecord.Value)
	} else {
		log.Printf("No update needed - zone: %s, name: %s", zoneID, configRecord.Name)
	}

	return nil
}
