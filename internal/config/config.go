package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL       string `json:"db_url"`
	CurrentUser string `json:"current_user"`
}

func Read() (Config, error) {
	filePath, err := getFilePath()
	if err != nil {
		log.Println(err.Error())
		return Config{}, err
	}
	fmt.Println(filePath)
	configFile, err := os.Open(filePath)
	if err != nil {
		log.Println(err.Error())
		return Config{}, err
	}
	defer configFile.Close()
	bytes, err := io.ReadAll(configFile)
	if err != nil {
		log.Println(err.Error())
		return Config{}, err
	}
	cfg := Config{}
	err = json.Unmarshal(bytes, &cfg)
	if err != nil {
		log.Println(err.Error())
		return Config{}, err
	}

	return cfg, nil
}

func (cfg *Config) SetUser(user string) error {
	cfg.CurrentUser = user
	err := cfg.write()
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (cfg *Config) write() error {
	filePath, err := getFilePath()
	if err != nil {
		log.Println(err)
		return err
	}

	b, err := json.Marshal(cfg)
	if err != nil {
		log.Println(err)
		return err
	}

	err = os.WriteFile(filePath, b, 0644)
	return err
}

func getFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	result := filepath.Join(homeDir, configFileName)
	return result, nil
}
