package util

import (
	"fmt"
	"time"
)

const (
	dateFormat         = "2006-01-02"
	dateFormatDE       = "02.01.2006"
	dateFormatDEShort  = "02.01."
	dateTimeFormat     = "2006-01-02T15:04:05.999999999"
	dateTimeFormatForm = "02.01.2006 15:04"
	timeFormat         = "15:04"
)

func ParseDateTime(dateTime string) (*time.Time, error) {
	t, err := time.Parse(dateTimeFormat, dateTime)
	if err != nil {
		return nil, fmt.Errorf("could not parse date time from '%s'", dateTime)
	}
	return &t, nil
}

func ParseDateTimeForm(dateTime string) (*time.Time, error) {
	t, err := time.Parse(dateTimeFormatForm, dateTime)
	if err != nil {
		return nil, fmt.Errorf("could not parse date time from '%s'", dateTime)
	}
	return &t, nil
}

func FormatTime(dateTime time.Time) string {
	return dateTime.Format(timeFormat)
}

func FormatDateTime(dateTime time.Time) string {
	return dateTime.Format(dateTimeFormat)
}

func FormatDateDE(dateTime time.Time) string {
	return dateTime.Format(dateFormatDE)
}

func FormatDateDEShort(dateTime time.Time) string {
	return dateTime.Format(dateFormatDEShort)
}

func ParseDate(date string) (*time.Time, error) {
	t, err := time.Parse(dateFormat, date)
	if err != nil {
		return nil, fmt.Errorf("could not parse date from '%s'", date)
	}
	return &t, nil
}
