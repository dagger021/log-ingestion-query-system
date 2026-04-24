CREATE TABLE IF NOT EXISTS logEntries (
    id String,

    level LowCardinality(String),
    message String,
    resourceId String,
    timestamp DateTime64(3),
    traceId String,
    spanId String,
    commit String,
    parentResourceId String
)
ENGINE = ReplacingMergeTree()
PARTITION BY toDate(timestamp)
ORDER BY (timestamp, resourceId, traceId, spanId, id);

-- ====================
-- Add Indexes
-- ====================
ALTER TABLE logEntries DROP INDEX IF EXISTS idx_level;
ALTER TABLE logEntries
ADD INDEX idx_level level TYPE set(100) GRANULARITY 1;

ALTER TABLE logEntries DROP INDEX IF EXISTS idx_resource;
ALTER TABLE logEntries
ADD INDEX idx_resource resourceId TYPE set(1000) GRANULARITY 1;

ALTER TABLE logEntries DROP INDEX IF EXISTS idx_trace;
ALTER TABLE logEntries
ADD INDEX idx_trace traceId TYPE bloom_filter(0.01) GRANULARITY 1;

ALTER TABLE logEntries DROP INDEX IF EXISTS idx_message;
ALTER TABLE logEntries
ADD INDEX idx_message message TYPE tokenbf_v1(1024, 3, 0) GRANULARITY 1;