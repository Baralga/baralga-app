package util

import (
	"fmt"
	"time"
)

const (
	dateFormat     = "2006-01-02"
	dateTimeFormat = "2006-01-02T15:04:05.999999999"
)

func ParseDateTime(dateTime string) (*time.Time, error) {
	t, err := time.Parse(dateTimeFormat, dateTime)
	if err != nil {
		return nil, fmt.Errorf("could not parse date time from '%s'", dateTime)
	}
	return &t, nil
}

func FormatDateTime(dateTime *time.Time) string {
	return dateTime.Format(dateTimeFormat)
}

func ParseDate(date string) (*time.Time, error) {
	t, err := time.Parse(dateFormat, date)
	if err != nil {
		return nil, fmt.Errorf("could not parse date from '%s'", date)
	}
	return &t, nil
}
