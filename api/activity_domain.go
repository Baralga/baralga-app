package main

import (
	"fmt"
	"time"

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
