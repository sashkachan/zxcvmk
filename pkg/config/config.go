package config

import (
	"encoding/json"
	"os"

	// "github.com/jedib0t/go-pretty/v6/table"
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

// Output outputs the given data into a supported format
func Output(data any, output string) (string, error) {
	switch output {
	case "json":
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", err
		}
		return string(jsonData), nil

	case "yaml":
		yamlData, err := yaml.Marshal(data)
		if err != nil {
			return "", err
		}
		return string(yamlData), nil

	case "table":
		// TODO: table export
		return "", nil

	}
	return "", nil
}
