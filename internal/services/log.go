package services

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/dagger021/log-ingestion-query-system/internal/domain"
	"github.com/dagger021/log-ingestion-query-system/internal/kafka"
)

type LogEntryService interface {
	// GetLogs returns all logs from the storage.
	GetLogs(context.Context) ([]domain.LogEntry, error)

	// StoreLog stores the logEntry into the storage.
	StoreLog(context.Context, domain.LogEntry) error
}

type logEntryService struct {
	chConn   clickhouse.Conn
	producer kafka.Producer
}

const (
	logColumns         = "id, level, message, resourceId, timestamp, traceId, spanId, commit, parentResourceId"
	baseLogEntrySelect = "SELECT " + logColumns + " FROM log_entries"
)

// GetLogs implements [LogEntryService].
func (s *logEntryService) GetLogs(c context.Context) ([]domain.LogEntry, error) {
	var logEntries []domain.LogEntry
	if err := s.chConn.Select(c, &logEntries, baseLogEntrySelect); err != nil {
		return nil, err
	}

	return logEntries, nil
}

// StoreLog implements [LogEntryService].
func (s *logEntryService) StoreLog(c context.Context, logEntry domain.LogEntry) error {
	if err := s.producer.ProduceLog(c, logEntry); err != nil {
		return err
	}
	return nil
}

func NewLogEntryService(chConn clickhouse.Conn) LogEntryService {
	return &logEntryService{chConn: chConn}
}
