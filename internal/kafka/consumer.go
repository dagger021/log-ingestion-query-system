package kafka

import (
	"context"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/dagger021/log-ingestion-query-system/internal/config"
	"github.com/dagger021/log-ingestion-query-system/internal/domain"
	"github.com/dagger021/log-ingestion-query-system/pkg/retry"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Consumer interface {
	Run(context.Context)
}

type batchItem struct {
	log domain.LogEntry
	msg kafka.Message
}

type consumer struct {
	logger *zap.Logger

	producer Producer // for DLQ -> producer.dlq
	reader   *kafka.Reader
	chConn   clickhouse.Conn
}

func NewConsumer(brokers []string, chConn clickhouse.Conn, producer Producer, logger *zap.Logger) Consumer {
	return &consumer{
		logger: logger,

		producer: producer,

		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   string(LogsTopic),
			GroupID: config.KafkaGroupID,
		}),

		chConn: chConn,
	}
}

func (c *consumer) Close() {
	if err := c.reader.Close(); err != nil {
		c.logger.Error("error closing consumer", zap.Error(err))
		return
	}

	c.logger.Info("consumer closed")
}

func (c *consumer) Run(ctx context.Context) {
	defer c.Close()

	c.logger.Info("initiating")

	flushTicker := time.NewTicker(config.FlushInterval)
	defer flushTicker.Stop()

	batch := make([]batchItem, 0, config.MaxBatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		if err := c.flushBatch(ctx, batch); err != nil {
			// send all messages to DLQ
			for _, item := range batch {
				if err :=
					c.producer.SendToDLQ(ctx, item.msg.Key, item.msg.Value, err); err != nil {
					// log -> DLQ failed
					c.logger.Error("logs DLQ failed", zap.Error(err))
				}

			}

			return // Not commiting -> retry later
		}

		// commit success message
		msgs := make([]kafka.Message, len(batch))
		for i, item := range batch {
			msgs[i] = item.msg
		}

		if err := c.reader.CommitMessages(ctx, msgs...); err != nil {
			// commit failure -> don't clear batch
			c.logger.Error("failed committing messages", zap.Error(err))
			return
		}

		batch = batch[:0] // clear batch
	}

	for {
		select {
		case <-ctx.Done():
			// flush batch & stop consumer
			flush()
			return

		case <-flushTicker.C:
			// flush batch & continue
			flush()
		}

		// Add log to batch, and flush it if batch is populated to MaxBatchSize
		ctxFetch, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		msg, err := c.reader.FetchMessage(ctxFetch)
		cancel()
		if err != nil {
			continue
		}

		logEntry, err := domain.ParseLogFromJSON(msg.Value)
		if err != nil {
			// skip poison message with commit
			c.reader.CommitMessages(ctx, msg)
			continue
		}

		// append log to the batch
		batch = append(batch, batchItem{log: *logEntry, msg: msg})

		if len(batch) >= config.MaxBatchSize {
			// batch is full
			flush()

			continue
		}
	}
}

func (c *consumer) flushBatch(ctx context.Context, batch []batchItem) error {
	return retry.Do(10, 500*time.Millisecond, func() error {
		chBatch, err := c.chConn.PrepareBatch(ctx, flushQuery)
		if err != nil {
			return err
		}

		for _, item := range batch {
			if err := chBatch.Append(
				item.log.ID,
				item.log.Level,
				item.log.Message,
				item.log.ResourceID,
				item.log.Timestamp,
				item.log.TraceID,
				item.log.SpanID,
				item.log.Commit,
				item.log.Metadata.ParentResourceID,
			); err != nil {
				return err
			}
		}

		return chBatch.Send()
	})
}

const flushQuery = `
	INSERT INTO log_entries (
		id,
		level,
		message,
		resource_id,
		timestamp,
		trace_id,
		span_id,
		commit,
		parent_resource_id
	)`
