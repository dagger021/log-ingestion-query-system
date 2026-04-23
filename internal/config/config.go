package config

import "time"

const (
	KafkaBroker  = "kafka:9092"
	KafkaGroupID = "log-ingestor"

	ClickhouseDSN = "clickhouse://@localhost:9000/default"

	BatchSize     = 1000
	FlushInterval = 2 * time.Second // flush batches in every two seconds, if batch is not full
)
