package seasons

import "errors"

var (
	ErrSeasonNotFoundForDate  = errors.New("season not found for date")
	ErrMultipleSeasonsForDate = errors.New("multiple seasons found for date")
)
