package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type LogEntry struct {
	ID         uuid.UUID `json:"id"`
	Level      LogLevel  `json:"level"`
	Message    string    `json:"message"`
	ResourceID string    `json:"resourceId"`
	Timestamp  time.Time `json:"timestamp"`
	TraceID    string    `json:"traceId"`
	SpanID     string    `json:"spanId"`
	Commit     string    `json:"commit"`

	Metadata struct {
		ParentResourceID string `json:"parentResourceId"`
	} `json:"metadata"`
}

func ParseLogFromJSON(data []byte) (*LogEntry, error) {
	var log LogEntry
	if err := json.Unmarshal(data, &log); err != nil {
		return nil, err
	}

	if _, ok := ToLogLevel(string(log.Level)); !ok {
		return nil, errors.New("invaild log level")
	}

	// populate log ID
	log.ID = uuid.New()
	return &log, nil
}

func (l *LogEntry) ToJSON() ([]byte, error) {
	data, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}

	return data, nil
}
