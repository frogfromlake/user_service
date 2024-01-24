package util

import (
	"fmt"
	"time"

	"github.com/google/uuid"
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

func ConvertToTimestamp(value time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: value, Valid: true}
}

func ConvertToTimestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value, Valid: true}
}

func ConvertToUUID(value string) pgtype.UUID {
	uuidValue, err := uuid.Parse(value)
	if err != nil {
		fmt.Printf("error: ConvertToUUID: failed to parse UUID: %v", err)
		return pgtype.UUID{Valid: false}
	}

	var uuidBytes [16]byte
	copy(uuidBytes[:], uuidValue[:])
	return pgtype.UUID{Bytes: uuidBytes, Valid: true}
}
