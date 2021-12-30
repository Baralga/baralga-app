package main

import (
	"fmt"
	"math"
	"time"

	"github.com/baralga/util"
	"github.com/google/uuid"
)

// Activity represents a tracked time for a project
type Activity struct {
	ID             uuid.UUID
	Start          time.Time
	End            time.Time
	Description    string
	ProjectID      uuid.UUID
	OrganizationID uuid.UUID
	Username       string
}

// ActivityFilter reprensents a filter for activities
type ActivityFilter struct {
	Timespan string
	start    time.Time
	end      time.Time
}

// Timespans for activity filter
const (
	TimespanYear    string = "year"
	TimespanQuarter string = "quarter"
	TimespanMonth   string = "month"
	TimespanWeek    string = "week"
	TimespanDay     string = "day"
	TimespanCustom  string = "custom"
)

// Start returns the filter's start date
func (f *ActivityFilter) Start() time.Time {
	return f.start
}

// End returns the filter's end date
func (f *ActivityFilter) End() time.Time {
	switch f.Timespan {
	case TimespanCustom:
		return f.end
	case TimespanDay:
		return f.start.AddDate(0, 0, 1)
	case TimespanWeek:
		return f.start.AddDate(0, 0, 7)
	case TimespanMonth:
		return f.start.AddDate(0, 1, 0)
	case TimespanQuarter:
		return f.start.AddDate(0, 3, 0)
	case TimespanYear:
		return f.start.AddDate(1, 0, 0)
	default:
		return f.end
	}
}

func (f *ActivityFilter) Next() *ActivityFilter {
	nextFilter := &ActivityFilter{
		Timespan: f.Timespan,
		start:    f.start,
		end:      f.end,
	}

	switch nextFilter.Timespan {
	case TimespanDay:
		nextFilter.start = f.start.AddDate(0, 0, 1)
		nextFilter.end = f.end.AddDate(0, 0, 1)
	case TimespanWeek:
		nextFilter.start = f.start.AddDate(0, 0, 7)
		nextFilter.end = f.end.AddDate(0, 0, 7)
	case TimespanMonth:
		nextFilter.start = f.start.AddDate(0, 1, 0)
		nextFilter.end = f.end.AddDate(0, 1, 0)
	case TimespanQuarter:
		nextFilter.start = f.start.AddDate(0, 3, 0)
		nextFilter.end = f.end.AddDate(0, 3, 0)
	case TimespanYear:
		nextFilter.start = f.start.AddDate(1, 0, 0)
		nextFilter.end = f.end.AddDate(1, 0, 0)
	}

	return nextFilter
}

func (f *ActivityFilter) Previous() *ActivityFilter {
	previousFilter := &ActivityFilter{
		Timespan: f.Timespan,
		start:    f.start,
		end:      f.end,
	}

	switch previousFilter.Timespan {
	case TimespanDay:
		previousFilter.start = f.start.AddDate(0, 0, -1)
		previousFilter.end = f.end.AddDate(0, 0, -1)
	case TimespanWeek:
		previousFilter.start = f.start.AddDate(0, 0, -7)
		previousFilter.end = f.end.AddDate(0, 0, -7)
	case TimespanMonth:
		previousFilter.start = f.start.AddDate(0, -1, 0)
		previousFilter.end = f.end.AddDate(0, -1, 0)
	case TimespanQuarter:
		previousFilter.start = f.start.AddDate(0, -3, 0)
		previousFilter.end = f.end.AddDate(0, -3, 0)
	case TimespanYear:
		previousFilter.start = f.start.AddDate(-1, 0, 0)
		previousFilter.end = f.end.AddDate(-1, 0, 0)
	}

	return previousFilter
}

// End returns the filter's display name
func (f *ActivityFilter) String() string {
	switch f.Timespan {
	case TimespanCustom:
		return f.Start().Format("2006-01-02") + "_" + f.End().Format("2006-01-02")
	case TimespanDay:
		return f.Start().Format("2006-01-02")
	case TimespanWeek:
		y, w := f.Start().ISOWeek()
		return fmt.Sprintf("%v-%v", y, w)
	case TimespanMonth:
		return f.Start().Format("2006-01")
	case TimespanQuarter:
		q := int(math.Ceil(float64(f.Start().Month()) / 3))
		return fmt.Sprintf("%v-%v", f.Start().Format("2006"), q)
	case TimespanYear:
		return f.Start().Format("2006")
	default:
		return "Custom"
	}
}

func (f *ActivityFilter) StringFormatted() string {
	switch f.Timespan {
	case TimespanCustom:
		return fmt.Sprintf(
			"%v - %v",
			util.FormatDateDEShort(f.Start()),
			util.FormatDateDEShort(f.End()),
		)
	case TimespanDay:
		return util.FormatDateDEShort(f.Start())
	default:
		return fmt.Sprintf(
			"%v - %v",
			util.FormatDateDEShort(f.Start()),
			util.FormatDateDEShort(f.End().AddDate(0, 0, -1)),
		)
	}
}

func (f *ActivityFilter) NewValue() string {
	now := time.Now()
	switch f.Timespan {
	case TimespanDay:
		return now.Format("2006-01-02")
	case TimespanWeek:
		y, w := now.ISOWeek()
		return fmt.Sprintf("%v-%v", y, w)
	case TimespanMonth:
		return now.Format("2006-01")
	case TimespanQuarter:
		q := int(now.Month()) / 3
		return fmt.Sprintf("%v-%v", now.Format("2006"), q)
	case TimespanYear:
		return now.Format("2006")
	default:
		return "Custom"
	}
}

// DurationHours is the activity duration in hours (e.g. 3)
func (ad *Activity) DurationHours() int {
	return int(ad.duration().Hours())
}

// DurationMinutes is the activity duration in minutes of unfinished hour (e.g. 15)
func (ad *Activity) DurationMinutes() int {
	m := int(ad.duration().Minutes())
	return m % 60
}

// DurationDecimal is the activity duration as decimal (e.g. 0.75)
func (ad *Activity) DurationDecimal() float64 {
	return ad.duration().Minutes() / 60.0
}

// DurationDecimal is the activity duration as formatted string (e.g. 1:15 h)
func (ad *Activity) DurationFormatted() string {
	return fmt.Sprintf("%v:%02d h", ad.DurationHours(), ad.DurationMinutes())
}

func (a *Activity) duration() time.Duration {
	return a.End.Sub(a.Start)
}
