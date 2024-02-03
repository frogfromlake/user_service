package util

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func ConvertToInt4(value int32) pgtype.Int4 {
	return pgtype.Int4{Int32: value, Valid: true}
}

func ConvertToInt8(value int64) pgtype.Int8 {
	return pgtype.Int8{Int64: value, Valid: true}
}

func ConvertToText(value string) pgtype.Text {
	return pgtype.Text{String: value, Valid: true}
}

func ConvertToDate(value time.Time) pgtype.Date {
	return pgtype.Date{Time: value, Valid: true}
}
