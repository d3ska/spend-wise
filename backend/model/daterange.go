package model

import (
	"errors"
	"time"
)

// DateRange represents an immutable half-open time interval [from, to).
// Fields are unexported to enforce immutability.
type DateRange struct {
	from time.Time
	to   time.Time
}

// NewDateRange creates a DateRange. Returns an error if from >= to.
func NewDateRange(from, to time.Time) (DateRange, error) {
	if !from.Before(to) {
		return DateRange{}, errors.New("daterange: from must be strictly before to")
	}
	return DateRange{from: from, to: to}, nil
}

// MonthRange returns a DateRange spanning the entire month (1st of month to 1st of next month).
func MonthRange(year int, month int) DateRange {
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	return DateRange{from: from, to: to}
}

// WeekRange returns a DateRange spanning ISO week (Monday to next Monday).
func WeekRange(year int, week int) DateRange {
	// Find January 4th of the year (always in ISO week 1).
	jan4 := time.Date(year, time.January, 4, 0, 0, 0, 0, time.UTC)
	// Find the Monday of ISO week 1.
	_, isoWeek := jan4.ISOWeek()
	monday := jan4.AddDate(0, 0, -int(jan4.Weekday()-time.Monday))
	if jan4.Weekday() == time.Sunday {
		monday = jan4.AddDate(0, 0, -6)
	}
	// Offset to the desired week.
	from := monday.AddDate(0, 0, (week-isoWeek)*7)
	to := from.AddDate(0, 0, 7)
	return DateRange{from: from, to: to}
}

// DayRange returns a DateRange spanning a single day [date, date+1day).
func DayRange(date time.Time) DateRange {
	from := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	to := from.AddDate(0, 0, 1)
	return DateRange{from: from, to: to}
}

// From returns the start of the range.
func (dr DateRange) From() time.Time { return dr.from }

// To returns the end of the range (exclusive).
func (dr DateRange) To() time.Time { return dr.to }
