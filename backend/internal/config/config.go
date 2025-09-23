package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/wb-go/wbf/zlog"
)

// Config holds the main configuration for the application.
type Config struct {
	Server   Server   `mapstructure:"server"`
	Database Database `mapstructure:"database"`
	Storage  Storage  `mapstructure:"storage"`
	Kafka    Kafka    `mapstructure:"kafka"`
	Retry    Retry    `mapstructure:"retry"`
}

// Server holds HTTP server-related configuration.
type Server struct {
	HTTPPort string `mapstructure:"http_port"` // HTTP port to listen on
}

// Database holds database master and slave configuration.
type Database struct {
	Master DatabaseNode   `mapstructure:"master"`
	Slaves []DatabaseNode `mapstructure:"slaves"`

	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// DatabaseNode holds connection parameters for a single database node.
type DatabaseNode struct {
	Host    string `mapstructure:"host"`
	Port    string `mapstructure:"port"`
	User    string `mapstructure:"user"`
	Pass    string `mapstructure:"pass"`
	Name    string `mapstructure:"name"`
	SSLMode string `mapstructure:"ssl_mode"`
}

// Storage holds configuration for the file storage backend.
type Storage struct {
	Endpoint   string `mapstructure:"endpoint"`
	AccessKey  string `mapstructure:"access_key"`
	SecretKey  string `mapstructure:"secret_key"`
	BucketName string `mapstructure:"bucket_name"`
	UseSSL     bool   `mapstructure:"use_ssl"`
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

// DSN returns the PostgreSQL DSN string for connecting to this database node.
func (n DatabaseNode) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		n.User, n.Pass, n.Host, n.Port, n.Name, n.SSLMode,
	)
}

// mustBindEnv binds critical environment variables to Viper keys.
//
// It panics if any environment variable cannot be bound.
func mustBindEnv() {
	bindings := map[string]string{
		"database.master.host": "DB_HOST",
		"database.master.port": "DB_PORT",
		"database.master.user": "DB_USER",
		"database.master.pass": "DB_PASSWORD",
		"database.master.name": "DB_NAME",
	}

	for key, env := range bindings {
		if err := viper.BindEnv(key, env); err != nil {
			zlog.Logger.Panic().Err(err).Msgf("failed to bind env %s", env)
		}
	}
}

// MustLoad loads the configuration from the specified file path.
// It panics if the configuration file cannot be loaded or unmarshaled.
func MustLoad(path string) *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		zlog.Logger.Panic().Err(err).Msg("failed to read config")
	}

	mustBindEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		zlog.Logger.Panic().Err(err).Msgf("failed to unmarshal config: %v", err)
	}

	return &cfg
}
