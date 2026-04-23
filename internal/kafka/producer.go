package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/dagger021/log-ingestion-query-system/internal/domain"
	"github.com/dagger021/log-ingestion-query-system/pkg/retry"
	"github.com/segmentio/kafka-go"
)

type Producer interface {
	Close() error

	ProduceLog(c context.Context, logEntry domain.LogEntry) error

	SendToDLQ(c context.Context, key, value []byte, ogErr error) error
}

type producer struct {
	writer, dlq *kafka.Writer
}

func NewProducer(brokers []string) Producer {
	return &producer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    string(LogsTopic),
			Balancer: &kafka.Hash{},
		},
		dlq: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    string(DLQTopic),
			Balancer: &kafka.Hash{},
		},
	}
}

func (p *producer) Close() error {
	if err := p.writer.Close(); err != nil {
		return err
	}
	return p.dlq.Close()
}

func (p *producer) writeWithRetries(c context.Context, key, value []byte) error {
	if err := retry.Do(10, 200*time.Millisecond, func() error {
		return p.writer.WriteMessages(c, kafka.Message{Key: key, Value: value})
	}); err != nil {
		return p.SendToDLQ(c, (key), value, err)
	}

	return nil
}

func (p *producer) SendToDLQ(c context.Context, key, value []byte, ogErr error) error {
	msgBytes, err := json.Marshal(DLQMessage{
		EventID:   string(key),
		Payload:   value,
		Error:     ogErr.Error(),
		Timestamp: time.Now(),
	})

	if err != nil {
		return err
	}

	return p.dlq.WriteMessages(c, kafka.Message{
		Key:     key,
		Value:   msgBytes,
		Headers: []kafka.Header{{Key: "error", Value: []byte(ogErr.Error())}},
	})
}

func (p *producer) ProduceLog(c context.Context, logEntry domain.LogEntry) error {
	payload, err := logEntry.ToJSON()
	if err != nil {
		return err
	}

	return p.writeWithRetries(c, logEntry.ID[:], payload)
}
