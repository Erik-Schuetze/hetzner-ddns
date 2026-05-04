package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Record struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
	TTL  int    `yaml:"ttl"`
}

type Zone struct {
	Name    string   `yaml:"zone_name"`
	Records []Record `yaml:"records"`
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
	fmt.Printf("refresh_interval: %d\n", c.Params.RefreshInterval)
	for _, zone := range c.Hetzner.Zones {
		fmt.Printf("ZoneName: %s\n", zone.Name)
		for _, record := range zone.Records {
			fmt.Printf("  Type: %s, Name: %s, TTL: %d\n", record.Type, record.Name, record.TTL)
		}
	}
}
