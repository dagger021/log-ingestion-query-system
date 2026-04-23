CREATE TABLE IF NOT EXISTS logEntries (
    id String,

    level LowCardinality(String),
    message String,
    resourceId String,
    timestamp DateTime64(3),
    traceId String,
    spanId String,
    commit String,
    parent_resourceId String,
)
ENGINE = ReplacingMergeTree()
PARTITION BY toDate(timestamp)
ORDER BY (timestamp, resourceId, traceId, spanId, id);