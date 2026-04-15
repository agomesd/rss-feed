package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const CONFIG_FILE_NAME string = ".gatorconfig.json"

type Config struct {
	DBURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	config := Config{}
	filepath, err := getConfigFilePath()
	if err != nil {
		return config, err
	}

	file, err := os.Open(filepath)
	if err != nil {
		return config, err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.Decode(&config)
	return config, nil
}

func (c *Config) SetUser(user string) error {
	c.CurrentUserName = user

	err := write(*c)

	if err != nil {
		return err
	}

	return nil

}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	filepath := fmt.Sprintf("%s/%s", homeDir, CONFIG_FILE_NAME)
	return filepath, nil
}

func write(cfg Config) error {
	filepath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		return nil
	}

	return nil
}
