package config

import (
	"fmt"
	"os"
      	"strings"


	"gopkg.in/yaml.v3"

	"github.com/driftctl/driftctl/internal/model"
)

// File represents the driftctl.yaml configuration.
type File struct {
	Database string            `yaml:"database"`
	API      APIConfig         `yaml:"api"`
	Workspaces []model.Workspace `yaml:"workspaces"`
}

// APIConfig holds server settings.
type APIConfig struct {
	Addr   string `yaml:"addr"`
	APIKey string `yaml:"api_key"`
}

// Load reads configuration from a YAML file.
func Load(path string) (*File, error) {
        if strings.Contains(path, "..") {
		return nil, fmt.Errorf("invalid config path: must not contain '..'")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg File
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Database == "" {
		cfg.Database = "driftctl.db"
	}
	if cfg.API.Addr == "" {
		cfg.API.Addr = ":8080"
	}
	return &cfg, nil
}
