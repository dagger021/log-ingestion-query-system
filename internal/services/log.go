package services

import (
	"context"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/dagger021/log-ingestion-query-system/internal/domain"
	"github.com/dagger021/log-ingestion-query-system/internal/kafka"
	"github.com/dagger021/log-ingestion-query-system/internal/logger"
	"go.uber.org/zap"
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

	// time-range filter
	if f.FromTime != nil {
		query += " AND timestamp >= ?"
		args = append(args, *f.FromTime)
	}

	if f.ToTime != nil {
		query += " AND timestamp <= ?"
		args = append(args, *f.ToTime)
	}

	// multi filters
	if len(f.Levels) > 0 {
		placeholders := make([]string, len(f.Levels))
		for i, val := range f.Levels {
			placeholders[i] = "?"
			args = append(args, val)
		}
		query += " AND level IN (" + strings.Join(placeholders, ",") + ")"
	}

	if len(f.ResourceIds) > 0 {
		placeholders := make([]string, len(f.ResourceIds))
		for i, val := range f.ResourceIds {
			placeholders[i] = "?"
			args = append(args, val)
		}
		query += " AND resourceId IN (" + strings.Join(placeholders, ",") + ")"
	}

	if len(f.TraceIds) > 0 {
		placeholders := make([]string, len(f.TraceIds))
		for i, val := range f.TraceIds {
			placeholders[i] = "?"
			args = append(args, val)
		}
		query += " AND traceId IN (" + strings.Join(placeholders, ",") + ")"
	}

	// regex search on Message
	if f.MessageRegex != nil && *f.MessageRegex != "" {
		query += " AND match(message, ?)"
		args = append(args, *f.MessageRegex)
	} else {
		for _, token := range f.MessageTokens {
			query += " AND hasToken(message, ?)"
			args = append(args, token)
		}
	}

	query += " ORDER BY timestamp DESC" // latest logs first
	// pagination
	if f.Limit == 0 {
		f.Limit = 100 // default default
	}
	f.Limit = min(1000, f.Limit) // limit upper-bound to 1000

	query += " LIMIT ? OFFSET ?"
	args = append(args, f.Limit, f.Offset)

	log := logger.FromContext(c)
	log.Debug("select query to db", zap.Any("query", query), zap.Any("query args", args))
	var logEntriesDB []domain.LogEntryDB
	if err := s.chConn.Select(c, &logEntriesDB, query, args...); err != nil {
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
