package config

import (
	"encoding/json"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

func ConfigGlobalLogger(cfg *Config) {
	writers := []io.Writer{}
	for _, o := range cfg.Log.Output {
		switch o {
		case "file":
			lumberjack := &cfg.Log.Lumberjack
			writers = append(writers, lumberjack)
			lumberjack.Rotate()
		case "stdout":
			if os.Getenv("PRETTY_LOGGING") == "true" {
				writers = append(writers, zerolog.ConsoleWriter{Out: os.Stdout})
			} else {
				writers = append(writers, os.Stdout)
			}
		}
	}
	log.Logger = log.Output(zerolog.MultiLevelWriter(writers...))

	lvl := map[string]zerolog.Level{
		"trace":    zerolog.TraceLevel,
		"debug":    zerolog.DebugLevel,
		"info":     zerolog.InfoLevel,
		"warn":     zerolog.WarnLevel,
		"error":    zerolog.ErrorLevel,
		"":         zerolog.ErrorLevel, // default
		"fatal":    zerolog.FatalLevel,
		"panic":    zerolog.PanicLevel,
		"no":       zerolog.NoLevel,
		"disabled": zerolog.Disabled,
	}[cfg.Log.Level]
	log.Logger = log.Level(lvl)
}
