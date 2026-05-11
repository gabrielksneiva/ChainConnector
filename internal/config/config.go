package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// RPC Configuration
	EthRPCURL     string
	SepoliaRPCURL string

	// Database Configuration
	DatabaseURL string

	// Server Configuration
	HTTPPort string
	HTTPAddr string

	// Monitoring Configuration
	MonitorInterval time.Duration
	MonitorEnabled  bool
	DefaultChain    string

	// Logging Configuration
	LogLevel string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		EthRPCURL:       getEnv("ETH_RPC_URL", "http://localhost:8545"),
		SepoliaRPCURL:   getEnv("SEPOLIA_RPC_URL", "http://localhost:8545"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/chainconnector"),
		HTTPPort:        getEnv("HTTP_PORT", "3000"),
		HTTPAddr:        getEnv("HTTP_ADDR", "0.0.0.0:3000"),
		MonitorInterval: getDurationEnv("MONITOR_INTERVAL_SECONDS", 15*time.Second),
		MonitorEnabled:  getBoolEnv("MONITOR_ENABLED", true),
		DefaultChain:    getEnv("DEFAULT_CHAIN", "sepolia"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getBoolEnv(key string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		v, err := strconv.ParseBool(value)
		if err != nil {
			return defaultVal
		}
		return v
	}
	return defaultVal
}

func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		seconds, err := strconv.ParseInt(value, 10, 64)
		if err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultVal
}
