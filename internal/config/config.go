package config

import (
	"encoding/json"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Bot botConfig `json:"bot"`
	DB  dbConfig  `json:"db"`
	Log logConfig `json:"log"`
}

type botConfig struct {
	Token          string `json:"token"`
	CheckPeriodMin int    `json:"check_period_min"`
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
	defer file.Close()

	var cfg *Config
	err = json.NewDecoder(file).Decode(&cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Bot.CheckPeriodMin < 1 {
		cfg.Bot.CheckPeriodMin = 15
	}

	return cfg, nil
}
