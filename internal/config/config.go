package config

import (
	"encoding/json"
	"io"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Bot botConfig `json:"bot"`
	DB  dbConfig  `json:"db"`
	Log logConfig `json:"log"`
}

type botConfig struct {
	Token string `json:"token"`
}

type dbConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	SLL      string `json:"sll"`
}

type logConfig struct {
	Level      string            `json:"level"`
	Output     []string          `json:"output"`
	Lumberjack lumberjack.Logger `json:"lumberjack"`
}

func Load() (*Config, error) {
	cfgPath := "./config.json"
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		cfgPath = p
	}

	file, err := os.Open(cfgPath)
	if err != nil {
		return nil, err
	}

	cfgJson, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var cfg *Config
	return cfg, json.Unmarshal(cfgJson, &cfg)
}
