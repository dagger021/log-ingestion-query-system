package services

import (
	"context"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/dagger021/log-ingestion-query-system/internal/domain"
	"github.com/dagger021/log-ingestion-query-system/internal/kafka"
)

type LogEntryService interface {
	// GetLogs returns all logs from the storage.
	GetLogs(context.Context, GetLogsFilter) ([]domain.LogEntry, error)

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
func (s *logEntryService) GetLogs(
	c context.Context, f GetLogsFilter,
) ([]domain.LogEntry, error) {
	query := baseLogEntrySelect + " WHERE 1=1" // selects all
	args := []any{}

	if f.Level != nil {
		query += " AND level = ?"
		args = append(args, *f.Level)
	}

	if f.ResourceId != nil {
		query += " AND resourceId = ?"
		args = append(args, *f.ResourceId)
	}

	if f.TraceId != nil {
		query += " AND traceId = ?"
		args = append(args, *f.TraceId)
	}

	if f.FromTime != nil {
		query += " AND fromTime >= ?"
		args = append(args, *f.FromTime)
	}

	if f.ToTime != nil {
		query += " AND toTime <= ?"
		args = append(args, *f.ToTime)
	}

	query += " ORDER BY timestamp DESC" // latest logs first
	// pagination
	if f.Limit == 0 {
		f.Limit = 100 // default default
	}
	query += " LIMIT ? OFFSET ?"
	args = append(args, f.Limit, f.Offset)

	var logEntriesDB []domain.LogEntryDB
	if err := s.chConn.Select(c, &logEntriesDB, query); err != nil {
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
