package clickhouse

import (
	"context"
	"os"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
)

func Migrate(conn clickhouse.Conn, file string) error {
	schema, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	// multi-line query is not allowed in clickhouse go driver
	// divide sql file into separate queries
	for query := range strings.SplitSeq(string(schema), ";") {
		if query != "" {
			if err := conn.Exec(context.Background(), query); err != nil {
				return err
			}
		}
	}

	return nil
}
