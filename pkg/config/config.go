package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

func LoadConfig(filename string) (*Config, error) {
	// Read the contents of the file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Declare a Config variable
	var config Config

	// Unmarshal the YAML data into the Config struct
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

type BackupTarget struct {
	Location        string   `yaml:"location"`
	PreRestoreHook  []string `yaml:"pre-restore-hook"`
	PostRestoreHook []string `yaml:"post-restore-hook"`
}

type Config struct {
	BackupProvider  string           `yaml:"backupProvider"`
	BackupProviders []BackupProvider `yaml:"backupProviders"`
	MountCommand    string           `yaml:"mountCommand"`
	BackupTargets   []BackupTarget   `yaml:"backupTargets"`
}

// BackupProvider provides detailed information about a specific backup provider.
type BackupProvider struct {
	Name                             string   `yaml:"name"`
	SnapshotListCommand              []string `yaml:"snapshotListCommand"`
	BackupRepositoryPasswordLocation string   `yaml:"backupRepositoryPasswordLocation"`
	BackupRepository                 string   `yaml:"backupRepository"`
}
