package kafka

type Topic string

const (
	LogsTopic Topic = "logs"
	DLQTopic  Topic = "logs-dlq"
)
