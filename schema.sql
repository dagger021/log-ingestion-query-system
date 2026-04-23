CREATE TABLE logs (
    id String,

    level LowCardinality(String),
    message String,
    resource_id String,
    timestamp DateTime64(3),
    trace_id String,
    span_id String,
    commit String,
    parent_resource_id String,
)
ENGINE = ReplacingMergeTree()
PARTITION BY toDate(timestamp)
ORDER BY (timestamp, resource_id, trace_id, span_id, id);