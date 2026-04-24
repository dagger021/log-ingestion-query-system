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
	baseLogEntrySelect = "SELECT " + logColumns + " FROM logEntries"
)

// GetLogs implements [LogEntryService].
func (s *logEntryService) GetLogs(c context.Context) ([]domain.LogEntry, error) {
	var logEntriesDB []domain.LogEntryDB
	if err := s.chConn.Select(c, &logEntriesDB, baseLogEntrySelect); err != nil {
		return nil, err
	}

	logEntries := make([]domain.LogEntry, len(logEntriesDB))
	for i, l := range logEntriesDB {
		logEntry, err := l.ToDomain()
		if err != nil {
			return nil, err
		}
		logEntries[i] = *logEntry
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

func NewLogEntryService(chConn clickhouse.Conn, producer kafka.Producer) LogEntryService {
	return &logEntryService{chConn: chConn, producer: producer}
}
