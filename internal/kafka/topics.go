package kafka

import (
	"errors"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Topic string

const (
	LogsTopic Topic = "logs"
	DLQTopic  Topic = "logs-dlq"
)

func EnsureTopics(brokers []string, logger *zap.Logger) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	topics := []kafka.TopicConfig{
		{
			Topic:             string(LogsTopic),
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		{
			Topic:             string(DLQTopic),
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = conn.CreateTopics(topics...)
	if err != nil {
		if errors.Is(err, kafka.TopicAlreadyExists) {
			logger.Warn("topic creation skipped or failed", zap.Error(err))
			return nil
		}
		return err
	}

	logger.Info("kafka topics ensured",
		zap.String("logs", string(LogsTopic)),
		zap.String("dlq", string(DLQTopic)),
	)

	return nil
}
