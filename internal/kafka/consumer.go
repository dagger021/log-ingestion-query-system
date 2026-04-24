package kafka

import (
	"context"
	"runtime"
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
	Close()
}

type batchItem struct {
	log domain.LogEntry
	msg kafka.Message
}

type consumer struct {
	logger *zap.Logger
	chConn clickhouse.Conn

	producer Producer
	reader   *kafka.Reader

	itemsCh chan batchItem
}

func NewConsumer(
	brokers []string,
	chConn clickhouse.Conn,
	producer Producer,
	logger *zap.Logger,
) Consumer {
	return &consumer{
		logger:   logger,
		chConn:   chConn,
		producer: producer,
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   string(LogsTopic),
			GroupID: config.KafkaGroupID,
		}),

		itemsCh: make(chan batchItem, config.MaxBatchSize*10),
	}
}

// Close implements [Consumer].
func (c *consumer) Close() {
	if err := c.reader.Close(); err != nil {
		c.logger.Error("failed closing kafka reader", zap.Error(err))
		return
	}
	c.logger.Info("consumer closed")
}

// Run implements [Consumer].
func (c *consumer) Run(ctx context.Context) {
	defer c.Close()

	c.logger.Info("consumer started")

	//worker pool for ingestion
	workerCount := min(runtime.NumCPU()*2, 32)
	for range workerCount {
		go c.ingestWork(ctx)
	}

	// one flush coordinator
	c.flushCoordinator(ctx)
}

func (c *consumer) ingestWork(ctx context.Context) {
	for {
		msg, err := c.fetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}

		logEntry, err := domain.ParseLogFromJSON(msg.Value)
		if err != nil {
			_ = c.reader.CommitMessages(ctx, msg)
			continue
		}

		select {
		case c.itemsCh <- batchItem{log: *logEntry, msg: msg}:
		case <-ctx.Done():
			return
		}
	}
}

func (c *consumer) fetchMessage(ctx context.Context) (kafka.Message, error) {
	ctxFetch, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	return c.reader.FetchMessage(ctxFetch)
}

func (c *consumer) flushCoordinator(ctx context.Context) {
	flushTicker := time.NewTicker(config.FlushInterval)
	defer flushTicker.Stop()

	batch := make([]batchItem, 0, config.MaxBatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		if err := c.flushBatch(ctx, batch); err != nil {
			c.logger.Error("flush batch failed", zap.Error(err))

			for _, item := range batch {
				if dlqErr := c.producer.SendToDLQ(ctx, item.msg.Key, item.msg.Value, err); dlqErr != nil {
					c.logger.Error("DLQ send failed", zap.Error(err))
				}
			}
			return
		}

		msgs := make([]kafka.Message, len(batch))
		for i, item := range batch {
			msgs[i] = item.msg
		}

		if err := c.reader.CommitMessages(ctx, msgs...); err != nil {
			c.logger.Error("commit messages failed", zap.Error(err))
			return
		}

		// empty batch
		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return

		case item := <-c.itemsCh:
			batch = append(batch, item)
			if len(batch) >= config.MaxBatchSize {
				flush()
			}

		case <-flushTicker.C:
			flush()
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
			logEntryDB := item.log.ToDB()
			if err := chBatch.Append(
				logEntryDB.ID,
				logEntryDB.Level,
				logEntryDB.Message,
				logEntryDB.ResourceID,
				logEntryDB.Timestamp,
				logEntryDB.TraceID,
				logEntryDB.SpanID,
				logEntryDB.Commit,
				logEntryDB.ParentResourceID,
			); err != nil {
				return err
			}
		}

		return chBatch.Send()
	})
}

const flushQuery = `
INSERT INTO logEntries (
	id,
	level,
	message,
	resourceId,
	timestamp,
	traceId,
	spanId,
	commit,
	parentResourceId
)`
