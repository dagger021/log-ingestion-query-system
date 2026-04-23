package clickhouse

import (
	"context"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"
)

func Migrate(conn clickhouse.Conn, file string) error {
	schema, err := os.ReadFile(file)
	if err == nil {
		err = conn.Exec(context.Background(), string(schema))
	}

	return err
}
