package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	configPath, err := getUserHomeDir()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	var unmarshaled Config
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		return Config{}, err
	}

	return unmarshaled, nil
}

func (cfg *Config) SetUser(name string) error {
	cfg.CurrentUserName = name

	configPath, err := getUserHomeDir()
	if err != nil {
		return err
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func getUserHomeDir() (string, error) {
	source, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(source, ".gatorconfig.json")

	return configPath, nil
}
