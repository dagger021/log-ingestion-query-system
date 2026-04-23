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

type baseLogEntry struct {
	ID         uuid.UUID `json:"id"`
	Level      LogLevel  `json:"level"`
	Message    string    `json:"message"`
	ResourceID string    `json:"resourceId"`
	Timestamp  time.Time `json:"timestamp"`
	TraceID    string    `json:"traceId"`
	SpanID     string    `json:"spanId"`
	Commit     string    `json:"commit"`
}

type LogEntry struct {
	baseLogEntry
	Metadata Metadata `json:"metadata"`
}

type LogEntryDB struct {
	baseLogEntry
	ParentResourceID string `json:"parentResourceId"`
}

func (l *LogEntryDB) ToDomain() LogEntry {
	return LogEntry{
		baseLogEntry: l.baseLogEntry,
		Metadata:     Metadata{ParentResourceID: l.ParentResourceID},
	}
}
func (l *LogEntry) ToDB() LogEntryDB {
	return LogEntryDB{
		baseLogEntry:     l.baseLogEntry,
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

func (l *baseLogEntry) Validate() error {
	if _, ok := ToLogLevel(string(l.Level)); !ok {
		return errors.New("invaild log level")
	}
	return nil
}

func (l LogEntry) ToJSON() ([]byte, error) {
	return json.Marshal(l)
}
