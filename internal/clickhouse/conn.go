package clickhouse

import (
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/dagger021/log-ingestion-query-system/pkg/retry"
)

func InitConn(chDSN string) (clickhouse.Conn, error) {
	opts, err := clickhouse.ParseDSN(chDSN)
	if err != nil {
		return nil, err
	}

	var conn clickhouse.Conn
	if err := retry.Do(10, 200*time.Millisecond, func() error {
		conn, err = clickhouse.Open(opts)
		return err
	}); err != nil {
		return nil, err
	}

	return conn, nil
}
