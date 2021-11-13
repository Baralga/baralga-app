package main

import (
	"testing"
	"time"

	"github.com/matryer/is"
)

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
