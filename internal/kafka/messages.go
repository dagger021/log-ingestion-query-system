package kafka

import "time"

type DLQMessage struct {
	EventID   string    `json:"event_id"`
	Error     string    `json:"error"`
	Payload   []byte    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}
