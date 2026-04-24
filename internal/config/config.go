package config

import (
	"fmt"
	"os"
	"time"
)

type EnvConfig struct {
	EnvironmentMode, KafkaBroker, ClickhouseDSN string
}

const (
	// EnvironmentMode = "development"

	// KafkaBroker = "kafka:9092"

	// ClickhouseDSN = "clickhouse://@localhost:9000/default"

	LoggerKey = "logger"

	KafkaGroupID = "log-ingestor"

	MaxBatchSize  = 1000
	FlushInterval = 2 * time.Second // flush batches in every two seconds, if batch is not full
)

func (c *EnvConfig) IsDev() bool {
	return c.EnvironmentMode == "development"
}

func InitConfig() (*EnvConfig, error) {
	var cfg EnvConfig

	envVal, err := getEnv("ENVIRONMENT_MODE")
	cfg.KafkaBroker = envVal
	if err != nil {
		return nil, err
	}
	cfg.EnvironmentMode = envVal

	envVal, err = getEnv("KAFKA_BROKER")
	if err != nil {
		return nil, err
	}
	cfg.KafkaBroker = envVal

	envVal, err = getEnv("CLICKHOUSE_DSN")
	if err != nil {
		return nil, err
	}
	cfg.ClickhouseDSN = envVal

	return &cfg, nil
}

func getEnv(key string) (string, error) {
	if val := os.Getenv(key); val != "" {
		return val, nil
	}
	return "", fmt.Errorf("env variable %s not set", key)
}
