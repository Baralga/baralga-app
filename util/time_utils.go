package util

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	dateFormat         = "2006-01-02"
	dateFormatDE       = "02.01.2006"
	dateFormatDEShort  = "2.1."
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

func FormatDate(dateTime time.Time) string {
	return dateTime.Format(dateFormat)
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

func Quarter(time time.Time) int {
	return int(math.Ceil(float64(time.Month()) / 3))
}

func CompleteTimeValue(time string) string {
	completedTime := time

	completedTime = strings.Replace(completedTime, ",,", ":", -1)
	completedTime = strings.Replace(completedTime, "/", ":", -1)
	completedTime = strings.Replace(completedTime, ";", ",", -1)
	completedTime = strings.Replace(completedTime, ".", ":", -1)

	// Treat 11,25 as 11:15
	// Treat 11,75 as 11:45
	// Treat 11,5 and 11,50 as 11:30
	splittedTime := strings.Split(completedTime, ",")
	if strings.Contains(completedTime, ",") && len(splittedTime) >= 2 {
		hh := splittedTime[0]
		mm := splittedTime[1]
		if len(mm) < 2 {
			mm = mm + "0"
		}

		// Convert to float for calculation
		fm, err := strconv.ParseFloat(mm, 64)
		if err != nil {
			return time
		}

		// Convert from base100 to base60
		fm *= 0.6
		// Round to int
		m := math.Round(fm)
		mm = fmt.Sprintf("%02.0f", m)

		if len(hh) < 2 {
			hh = "0" + hh
		}
		completedTime = hh + ":" + mm
		return completedTime
	}

	if strings.Contains(completedTime, ":") {
		return completedTime
	}

	_, err := strconv.ParseInt(completedTime, 10, 32)
	if err != nil {
		return time
	}

	if len(completedTime) < 2 {
		completedTime = "0" + completedTime
	}

	if !strings.Contains(completedTime, ":") {
		completedTime += ":00"
	}

	return completedTime
}
