package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestActivityTimeReportItemAsTime(t *testing.T) {
	// Arrange
	is := is.New(t)
	reportItem := &ActivityTimeReportItem{
		Year:  2022,
		Month: 9,
		Day:   3,
	}

	// Act
	time := reportItem.AsTime()

	// Assert
	is.Equal(time.Year(), reportItem.Year)
	is.Equal(int(time.Month()), reportItem.Month)
	is.Equal(time.Day(), reportItem.Day)
}

func TestActivityDurationHours(t *testing.T) {
	is := is.New(t)

	a := &Activity{
		Start: time.Now(),
		End:   time.Now().Add(1 * time.Hour).Add(30 * time.Minute),
	}

	hours := a.DurationHours()
	is.Equal(hours, 1)
}

func TestActivityDurationMinutes(t *testing.T) {
	is := is.New(t)

	a := &Activity{
		Start: time.Now(),
		End:   time.Now().Add(1 * time.Hour).Add(30 * time.Minute),
	}

	minutes := a.DurationMinutes()
	is.Equal(minutes, 30)
}

func TestActivityDurationMinutesTotal(t *testing.T) {
	is := is.New(t)

	a := &Activity{
		Start: time.Now(),
		End:   time.Now().Add(1 * time.Hour).Add(30 * time.Minute),
	}

	minutes := a.DurationMinutesTotal()
	is.Equal(minutes, 90)
}

func TestActivityDurationDecimal(t *testing.T) {
	is := is.New(t)
	now := time.Now()

	a := &Activity{
		Start: now,
		End:   now.Add(1 * time.Hour).Add(30 * time.Minute),
	}

	decimalDuration := a.DurationDecimal()
	is.Equal(decimalDuration, 1.5)
}

func TestActivityDurationFormatted(t *testing.T) {
	is := is.New(t)

	a := &Activity{
		Start: time.Now(),
		End:   time.Now().Add(1 * time.Hour).Add(9 * time.Minute),
	}

	formatted := a.DurationFormatted()
	is.Equal(formatted, "1:09 h")
}

func TestActivityFilterEnd(t *testing.T) {
	is := is.New(t)

	var f *ActivityFilter

	start, _ := time.Parse(time.RFC3339, "2021-11-12T11:00:00.000Z")
	end, _ := time.Parse(time.RFC3339, "2021-11-12T11:30:00.000Z")

	t.Run("End with custom filter", func(t *testing.T) {
		f = &ActivityFilter{
			start:    start,
			end:      end,
			Timespan: TimespanCustom,
		}
		is.Equal(f.End(), end)
	})

	t.Run("End with year filter", func(t *testing.T) {
		expectedEnd := start.AddDate(1, 0, 0)
		f = &ActivityFilter{
			start:    start,
			Timespan: TimespanYear,
		}
		is.Equal(f.End(), expectedEnd)
	})

	t.Run("End with week filter", func(t *testing.T) {
		expectedEnd := start.AddDate(0, 0, 7)
		f = &ActivityFilter{
			start:    start,
			Timespan: TimespanWeek,
		}
		is.Equal(f.End(), expectedEnd)
	})

	t.Run("End with quarter filter", func(t *testing.T) {
		expectedEnd := start.AddDate(0, 3, 0)
		f = &ActivityFilter{
			start:    start,
			Timespan: TimespanQuarter,
		}
		is.Equal(f.End(), expectedEnd)
	})

	t.Run("End with month filter", func(t *testing.T) {
		expectedEnd := start.AddDate(0, 1, 0)
		f = &ActivityFilter{
			start:    start,
			Timespan: TimespanMonth,
		}
		is.Equal(f.End(), expectedEnd)
	})

	t.Run("End with day filter", func(t *testing.T) {
		expectedEnd := start.AddDate(0, 0, 1)
		f = &ActivityFilter{
			start:    start,
			Timespan: TimespanDay,
		}
		is.Equal(f.End(), expectedEnd)
	})
}

func TestActivityFilterSting(t *testing.T) {
	is := is.New(t)

	var f *ActivityFilter

	start, _ := time.Parse(time.RFC3339, "2021-11-12T11:00:00.000Z")
	end, _ := time.Parse(time.RFC3339, "2021-11-12T11:30:00.000Z")

	t.Run("String with custom filter", func(t *testing.T) {
		f = &ActivityFilter{
			start:    start,
			end:      end,
			Timespan: TimespanCustom,
		}
		is.Equal(f.String(), "2021-11-12_2021-11-12")
	})

	t.Run("String with year filter", func(t *testing.T) {
		f = &ActivityFilter{
			start:    start,
			Timespan: TimespanYear,
		}
		is.Equal(f.String(), "2021")
	})

	t.Run("String with week filter", func(t *testing.T) {
		f = &ActivityFilter{
			start:    start,
			Timespan: TimespanWeek,
		}
		is.Equal(f.String(), "2021-45")
	})

	t.Run("String with quarter filter", func(t *testing.T) {
		f = &ActivityFilter{
			start:    start,
			Timespan: TimespanQuarter,
		}
		is.Equal(f.String(), "2021-4")
	})

	t.Run("String with month filter", func(t *testing.T) {
		f = &ActivityFilter{
			start:    start,
			Timespan: TimespanMonth,
		}
		is.Equal(f.String(), "2021-11")
	})

	t.Run("String with day filter", func(t *testing.T) {
		f = &ActivityFilter{
			start:    start,
			Timespan: TimespanDay,
		}
		is.Equal(f.String(), "2021-11-12")
	})
}

func TestActivityFilterHome(t *testing.T) {
	is := is.New(t)

	start, _ := time.Parse(time.RFC3339, "2021-11-12T11:00:00.000Z")
	f := &ActivityFilter{
		start:    start,
		Timespan: TimespanYear,
	}

	t.Run("String with year filter", func(t *testing.T) {
		homeFilter := f.Home()
		is.Equal(homeFilter.String(), fmt.Sprintf("%v", time.Now().Year()))
	})

}
