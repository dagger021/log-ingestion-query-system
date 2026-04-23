package logger

import (
	"context"

	"github.com/dagger021/log-ingestion-query-system/internal/config"
	"go.uber.org/zap"
)

func New(c *config.EnvConfig) (*zap.Logger, error) {
	if c.IsDev() {
		return zap.NewDevelopment()
	}

	return zap.NewProduction()
}

func FromContext(c context.Context) *zap.Logger {
	if l, ok := c.Value(config.LoggerKey).(*zap.Logger); ok {
		return l
	}

	return zap.L() // instead global logger
}
