package model

import "time"

type CryptoFilter struct {
	Page      int
	Limit     int
	Symbol    string
	StartTime *time.Time
	EndTime   *time.Time
}
