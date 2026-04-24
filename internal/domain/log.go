package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Metadata struct {
	ParentResourceID string `json:"parentResourceId"`
}

type LogEntry struct {
	ID         uuid.UUID `json:"id"`
	Level      LogLevel  `json:"level"`
	Message    string    `json:"message"`
	ResourceID string    `json:"resourceId"`
	Timestamp  time.Time `json:"timestamp"`
	TraceID    string    `json:"traceId"`
	SpanID     string    `json:"spanId"`
	Commit     string    `json:"commit"`

	Metadata Metadata `json:"metadata"`
}

type LogEntryDB struct {
	ID               string    `json:"id" ch:"id"`
	Level            string    `json:"level" ch:"level"`
	Message          string    `json:"message" ch:"message"`
	ResourceID       string    `json:"resourceId" ch:"resourceId"`
	Timestamp        time.Time `json:"timestamp" ch:"timestamp"`
	TraceID          string    `json:"traceId" ch:"traceId"`
	SpanID           string    `json:"spanId" ch:"spanId"`
	Commit           string    `json:"commit" ch:"commit"`
	ParentResourceID string    `json:"parentResourceId" ch:"parentResourceId"`
}

func (l *LogEntryDB) ToDomain() (*LogEntry, error) {
	level, _ := ToLogLevel(l.Level)
	id, err := uuid.Parse(l.ID)
	if err != nil {
		return nil, err
	}
	return &LogEntry{
		ID:         id,
		Level:      level,
		Message:    l.Message,
		ResourceID: l.ResourceID,
		Timestamp:  l.Timestamp,
		TraceID:    l.TraceID,
		SpanID:     l.SpanID,
		Commit:     l.Commit,
		Metadata:   Metadata{ParentResourceID: l.ParentResourceID},
	}, nil
}

func (l *LogEntry) ToDB() LogEntryDB {
	return LogEntryDB{
		ID:               l.ID.String(),
		Level:            string(l.Level),
		Message:          l.Message,
		ResourceID:       l.ResourceID,
		Timestamp:        l.Timestamp,
		TraceID:          l.TraceID,
		SpanID:           l.SpanID,
		Commit:           l.Commit,
		ParentResourceID: l.Metadata.ParentResourceID,
	}
}

func ParseLogFromJSON(data []byte) (*LogEntry, error) {
	var log LogEntry
	if err := json.Unmarshal(data, &log); err != nil {
		return nil, err
	}

	if err := log.Validate(); err != nil {
		return nil, err
	}

	// populate log ID
	log.ID = uuid.New()
	return &log, nil
}

func (l *LogEntry) Validate() error {
	if _, ok := ToLogLevel(string(l.Level)); !ok {
		return errors.New("invaild log level")
	}
	return nil
}

func (l LogEntry) ToJSON() ([]byte, error) {
	return json.Marshal(l)
}
