package config

import(
	"os"
	"encoding/json"
	"path/filepath"
)

type Config struct {
	Db_url string
	Current_user_name string
}

const configFileName = ".gatorconfig.json"

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	filePath := filepath.Join(home, configFileName)
	return filePath, nil
}

func Read() (*Config, error) {
	filePath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	rawBody, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var c *Config

	err = json.Unmarshal(rawBody, &c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func write(cfg Config) error {
	filePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) SetUser(name string) error {
	c.Current_user_name = name
	err := write(*c)
	if err != nil {
		return err
	}
	return nil
} 