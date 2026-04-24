package services

import "time"

type GetLogsFilter struct {
	Levels      []string
	ResourceIds []string
	TraceIds    []string

	FromTime, ToTime *time.Time

	MessageRegex  *string
	MessageTokens []string

	Limit, Offset uint64
}
