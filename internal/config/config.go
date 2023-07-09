package config

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// JSON config structure
type Config struct {
	StoragePath        string        `json:"storagePath"`  // the path to the directory with the emails
	TTL                time.Duration `json:"ttl"`          // reference lifetime
	Addr               string        `json:"addr"`         // server address
	Port               int           `json:"port"`         // server port
	CleaningTime       time.Duration `json:"cleaningTime"` // the interval of clearing elements with expired ttl
	AuthorizationToken string        `json:"authorizationToken"`
	Mu                 sync.Mutex
}

type App struct {
	Config  Config
	Logger  *logrus.Logger
	LinkMap map[string]map[string]time.Time
}

// Convert json to structure
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
		LinkMap: make(map[string]map[string]time.Time),
	}
}
