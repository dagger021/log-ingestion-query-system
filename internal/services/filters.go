package services

import "time"

type GetLogsFilter struct {
	Level      *string
	ResourceId *string
	TraceId    *string

	FromTime, ToTime *time.Time

	Limit, Offset uint64
}
