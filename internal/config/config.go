package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/erik-schuetze/hetzner-ddns/internal/hetzner"
)

type Zone struct {
	ZoneID  string           `yaml:"zone_id"`
	Records []hetzner.Record `yaml:"records"`
}

type Config struct {
	Params struct {
		RefreshInterval int `yaml:"refresh_interval"`
	} `yaml:"params"`
	Hetzner struct {
		Zones []Zone `yaml:"zones"`
	} `yaml:"hetzner"`
}

func Load(configPath string) (*Config, error) {
	config := new(Config)
	//yamlFile, err := os.ReadFile("/config/config.yaml")
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := yaml.Unmarshal(yamlFile, config); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return config, nil
}

func (c *Config) Print() {
	fmt.Printf("refresh_interval: %d", c.Params.RefreshInterval)
	for _, zone := range c.Hetzner.Zones {
		fmt.Printf("ZoneID: %s\n", zone.ZoneID)
		for _, record := range zone.Records {
			fmt.Printf("  Type: %s, Name: %s, TTL: %d\n", record.Type, record.Name, record.TTL)
		}
	}
}
