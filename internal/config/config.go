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
	EthWSURL      string
	SepoliaWSURL  string

	// Database Configuration
	DatabaseURL string

	// Server Configuration
	HTTPPort    string
	HTTPAddr    string
	HTTPEnabled bool

	// Monitoring Configuration
	MonitorInterval time.Duration
	MonitorEnabled  bool
	DefaultChain    string
	MigrationsDir   string

	// AWS/LocalStack SQS Configuration
	AWSEndpointURL       string
	AWSRegion            string
	AWSAccessKeyID       string
	AWSSecretAccessKey   string
	SQSEnabled           bool
	SQSConsumerEnabled   bool
	NetworkQueueName     string
	NetworkQueueURL      string
	SQSWaitTimeSeconds   int32
	SQSVisibilityTimeout int32

	// Block monitoring queue configuration
	BlockProducerEnabled bool
	BlockConsumerEnabled bool
	BlockQueueName       string
	BlockQueueURL        string

	// Logging Configuration
	LogLevel string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		EthRPCURL:            getEnv("ETH_RPC_URL", "http://localhost:8545"),
		SepoliaRPCURL:        getEnv("SEPOLIA_RPC_URL", "http://localhost:8545"),
		EthWSURL:             getEnv("ETH_WS_URL", "ws://localhost:8545"),
		SepoliaWSURL:         getEnv("SEPOLIA_WS_URL", "ws://localhost:8545"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/chainconnector"),
		HTTPPort:             getEnv("HTTP_PORT", "3000"),
		HTTPAddr:             getEnv("HTTP_ADDR", "0.0.0.0:3000"),
		HTTPEnabled:          getBoolEnv("HTTP_ENABLED", true),
		MonitorInterval:      getDurationEnv("MONITOR_INTERVAL_SECONDS", 15*time.Second),
		MonitorEnabled:       getBoolEnv("MONITOR_ENABLED", true),
		DefaultChain:         getEnv("DEFAULT_CHAIN", "sepolia"),
		MigrationsDir:        getEnv("MIGRATIONS_DIR", "./migrations"),
		AWSEndpointURL:       getEnv("AWS_ENDPOINT_URL", ""),
		AWSRegion:            getEnv("AWS_REGION", "us-east-1"),
		AWSAccessKeyID:       getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey:   getEnv("AWS_SECRET_ACCESS_KEY", ""),
		SQSEnabled:           getBoolEnv("SQS_ENABLED", false),
		SQSConsumerEnabled:   getBoolEnv("SQS_CONSUMER_ENABLED", false),
		NetworkQueueName:     getEnv("NETWORK_QUEUE_NAME", "chainconnector-network-registrations"),
		NetworkQueueURL:      getEnv("NETWORK_QUEUE_URL", ""),
		SQSWaitTimeSeconds:   getInt32Env("SQS_WAIT_TIME_SECONDS", 10),
		SQSVisibilityTimeout: getInt32Env("SQS_VISIBILITY_TIMEOUT_SECONDS", 30),
		BlockProducerEnabled: getBoolEnv("BLOCK_PRODUCER_ENABLED", false),
		BlockConsumerEnabled: getBoolEnv("BLOCK_CONSUMER_ENABLED", false),
		BlockQueueName:       getEnv("BLOCK_QUEUE_NAME", "chainconnector-block-events"),
		BlockQueueURL:        getEnv("BLOCK_QUEUE_URL", ""),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
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

func getInt32Env(key string, defaultVal int32) int32 {
	if value, exists := os.LookupEnv(key); exists {
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err == nil {
			return int32(parsed)
		}
	}
	return defaultVal
}
