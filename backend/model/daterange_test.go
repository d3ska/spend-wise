package model

import (
	"testing"
	"time"
)

func date(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func TestNewDateRange_Valid(t *testing.T) {
	dr, err := NewDateRange(date(2026, 2, 1), date(2026, 3, 1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !dr.From().Equal(date(2026, 2, 1)) {
		t.Errorf("From = %v, want 2026-02-01", dr.From())
	}
	if !dr.To().Equal(date(2026, 3, 1)) {
		t.Errorf("To = %v, want 2026-03-01", dr.To())
	}
}

func TestNewDateRange_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		from, to time.Time
	}{
		{"from after to", date(2026, 3, 1), date(2026, 2, 1)},
		{"same date", date(2026, 2, 1), date(2026, 2, 1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDateRange(tt.from, tt.to)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestMonthRange(t *testing.T) {
	dr := MonthRange(2026, 2)
	if !dr.From().Equal(date(2026, 2, 1)) {
		t.Errorf("From = %v, want 2026-02-01", dr.From())
	}
	if !dr.To().Equal(date(2026, 3, 1)) {
		t.Errorf("To = %v, want 2026-03-01", dr.To())
	}
}

func TestMonthRange_December(t *testing.T) {
	dr := MonthRange(2026, 12)
	if !dr.From().Equal(date(2026, 12, 1)) {
		t.Errorf("From = %v, want 2026-12-01", dr.From())
	}
	if !dr.To().Equal(date(2027, 1, 1)) {
		t.Errorf("To = %v, want 2027-01-01", dr.To())
	}
}

func TestWeekRange(t *testing.T) {
	dr := WeekRange(2026, 7)
	// ISO week 7 of 2026: Monday 2026-02-09 to Monday 2026-02-16
	wantFrom := date(2026, 2, 9)
	wantTo := date(2026, 2, 16)
	if !dr.From().Equal(wantFrom) {
		t.Errorf("From = %v, want %v", dr.From(), wantFrom)
	}
	if !dr.To().Equal(wantTo) {
		t.Errorf("To = %v, want %v", dr.To(), wantTo)
	}
}

func TestWeekRange_Week1(t *testing.T) {
	dr := WeekRange(2026, 1)
	// ISO week 1 of 2026: Monday 2025-12-29 to Monday 2026-01-05
	wantFrom := date(2025, 12, 29)
	wantTo := date(2026, 1, 5)
	if !dr.From().Equal(wantFrom) {
		t.Errorf("From = %v, want %v", dr.From(), wantFrom)
	}
	if !dr.To().Equal(wantTo) {
		t.Errorf("To = %v, want %v", dr.To(), wantTo)
	}
}

func TestDayRange(t *testing.T) {
	dr := DayRange(date(2026, 2, 12))
	if !dr.From().Equal(date(2026, 2, 12)) {
		t.Errorf("From = %v, want 2026-02-12", dr.From())
	}
	if !dr.To().Equal(date(2026, 2, 13)) {
		t.Errorf("To = %v, want 2026-02-13", dr.To())
	}
}
