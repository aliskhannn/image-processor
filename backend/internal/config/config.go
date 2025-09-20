package config

import (
	"time"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/zlog"
)

// Config holds the main configuration for the application.
type Config struct {
	Server  Server  `mapstructure:"server"`
	Storage Storage `mapstructure:"storage"`
	Kafka   Kafka   `mapstructure:"kafka"`
	Retry   Retry   `mapstructure:"retry"`
}

// Server holds HTTP server-related configuration.
type Server struct {
	HTTPPort string `mapstructure:"http_port"` // HTTP port to listen on
}

// Storage holds configuration for the file storage backend.
type Storage struct {
	BaseDir string `mapstructure:"base_dir"` // Base directory for storing files
}

// Kafka holds configuration for the Kafka message queue.
type Kafka struct {
	GroupID string   `mapstructure:"group_id"` // Consumer group ID
	Topic   string   `mapstructure:"topic"`    // Kafka topic name
	Brokers []string `mapstructure:"brokers"`  // List of Kafka broker addresses
}

// Retry defines retry policy configuration.
type Retry struct {
	Attempts int           `mapstructure:"attempts"` // Number of retry attempts
	Delay    time.Duration `mapstructure:"delay"`    // Initial delay between retries
	Backoff  float64       `mapstructure:"backoff"`  // Backoff multiplier for delays
}

// MustLoad loads the configuration from the specified file path.
// It panics if the configuration file cannot be loaded or unmarshaled.
func MustLoad(path string) *Config {
	c := config.New()

	if err := c.Load(path); err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to load config")
	}

	var cfg Config
	if err := c.Unmarshal(&cfg); err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to unmarshal config")
	}

	return &cfg
}
