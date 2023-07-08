package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type Config struct {
	StoragePath string        `json:"storagePath"`
	TTL         time.Duration `json:"ttl"`
	Addr        string        `json:"addr"`
	Port        int           `json:"port"`
}

type App struct {
	Config  Config
	Logger  *logrus.Logger
	LinkMap map[string]string
}

func NewConfig(configPath string) *App {

	file, err := os.Open(configPath)
	if err != nil {
		logrus.Fatalf("config file does not exist: %s", configPath)
	}
	defer file.Close()

	var config Config
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		logrus.Fatalf("cannot read config: %s", err)
	}

	logger := logrus.New()

	return &App{
		Config:  config,
		Logger:  logger,
		LinkMap: make(map[string]string),
	}
}
