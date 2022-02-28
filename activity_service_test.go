package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestTimeReportsByDay(t *testing.T) {
	// Arrange
	is := is.New(t)

	activityRepository := NewInMemActivityRepository()
	a := &app{
		ActivityRepository: activityRepository,
	}

	start1, _ := time.Parse(time.RFC3339, "2021-01-01T10:00:00.000Z")
	end1, _ := time.Parse(time.RFC3339, "2021-01-01T11:00:00.000Z")

	start2, _ := time.Parse(time.RFC3339, "2021-01-08T10:00:00.000Z")
	end2, _ := time.Parse(time.RFC3339, "2021-01-08T11:00:00.000Z")

	activityRepository.activities = []*Activity{
		{
			Start: start1,
			End:   end1,
		},
		{
			Start: start2,
			End:   end2,
		},
	}

	principal := &Principal{}
	filter := &ActivityFilter{}

	// Act
	timeReports, err := a.TimeReports(context.Background(), principal, filter, "day")

	// Assert
	is.NoErr(err)
	is.Equal(len(timeReports), 2)

	day1 := timeReports[0]
	is.Equal(day1.Year, 2021)
	is.Equal(day1.Day, 1)
	is.Equal(day1.DurationInMinutesTotal, 60)

	day2 := timeReports[1]
	is.Equal(day2.Year, 2021)
	is.Equal(day2.Day, 8)
	is.Equal(day2.DurationInMinutesTotal, 60)
}

func TestTimeReportsByWeek(t *testing.T) {
	// Arrange
	is := is.New(t)

	activityRepository := NewInMemActivityRepository()
	a := &app{
		ActivityRepository: activityRepository,
	}

	start1, _ := time.Parse(time.RFC3339, "2021-01-01T10:00:00.000Z")
	end1, _ := time.Parse(time.RFC3339, "2021-01-01T11:00:00.000Z")

	start2, _ := time.Parse(time.RFC3339, "2021-01-08T10:00:00.000Z")
	end2, _ := time.Parse(time.RFC3339, "2021-01-08T11:00:00.000Z")

	activityRepository.activities = []*Activity{
		{
			Start: start1,
			End:   end1,
		},
		{
			Start: start2,
			End:   end2,
		},
	}

	principal := &Principal{}
	filter := &ActivityFilter{}

	// Act
	timeReports, err := a.TimeReports(context.Background(), principal, filter, "week")

	// Assert
	is.NoErr(err)
	is.Equal(len(timeReports), 2)

	week1 := timeReports[0]
	is.Equal(week1.Year, 2021)
	is.Equal(week1.Week, 53)
	is.Equal(week1.DurationInMinutesTotal, 60)

	week2 := timeReports[1]
	is.Equal(week2.Year, 2021)
	is.Equal(week2.Week, 1)
	is.Equal(week2.DurationInMinutesTotal, 60)
}

func TestTimeReportsByMonth(t *testing.T) {
	// Arrange
	is := is.New(t)

	activityRepository := NewInMemActivityRepository()
	a := &app{
		ActivityRepository: activityRepository,
	}

	start1, _ := time.Parse(time.RFC3339, "2021-01-01T10:00:00.000Z")
	end1, _ := time.Parse(time.RFC3339, "2021-01-01T11:00:00.000Z")

	start2, _ := time.Parse(time.RFC3339, "2021-02-01T10:00:00.000Z")
	end2, _ := time.Parse(time.RFC3339, "2021-02-01T11:00:00.000Z")

	activityRepository.activities = []*Activity{
		{
			Start: start1,
			End:   end1,
		},
		{
			Start: start2,
			End:   end2,
		},
	}

	principal := &Principal{}
	filter := &ActivityFilter{}

	// Act
	timeReports, err := a.TimeReports(context.Background(), principal, filter, "month")

	// Assert
	is.NoErr(err)
	is.Equal(len(timeReports), 2)

	monthJan := timeReports[0]
	is.Equal(monthJan.Year, 2021)
	is.Equal(monthJan.Month, 1)
	is.Equal(monthJan.DurationInMinutesTotal, 60)

	monthFeb := timeReports[1]
	is.Equal(monthFeb.Year, 2021)
	is.Equal(monthFeb.Month, 2)
	is.Equal(monthFeb.DurationInMinutesTotal, 60)
}

func TestTimeReportsByQuarter(t *testing.T) {
	// Arrange
	is := is.New(t)

	activityRepository := NewInMemActivityRepository()
	a := &app{
		ActivityRepository: activityRepository,
	}

	start1, _ := time.Parse(time.RFC3339, "2021-01-01T10:00:00.000Z")
	end1, _ := time.Parse(time.RFC3339, "2021-01-01T11:00:00.000Z")

	start2, _ := time.Parse(time.RFC3339, "2021-04-01T10:00:00.000Z")
	end2, _ := time.Parse(time.RFC3339, "2021-04-01T11:00:00.000Z")

	activityRepository.activities = []*Activity{
		{
			Start: start1,
			End:   end1,
		},
		{
			Start: start2,
			End:   end2,
		},
	}

	principal := &Principal{}
	filter := &ActivityFilter{}

	// Act
	timeReports, err := a.TimeReports(context.Background(), principal, filter, "quarter")

	// Assert
	is.NoErr(err)
	is.Equal(len(timeReports), 2)

	q1 := timeReports[0]
	is.Equal(q1.Year, 2021)
	is.Equal(q1.Quarter, 1)
	is.Equal(q1.DurationInMinutesTotal, 60)

	q2 := timeReports[1]
	is.Equal(q2.Year, 2021)
	is.Equal(q2.Quarter, 2)
	is.Equal(q2.DurationInMinutesTotal, 60)
}

func TestWriteAsCSV(t *testing.T) {
	is := is.New(t)

	a := &app{}

	start, _ := time.Parse(time.RFC3339, "2021-11-12T11:00:00.000Z")
	end, _ := time.Parse(time.RFC3339, "2021-11-12T11:30:00.000Z")

	activity := &Activity{
		Start:     start,
		End:       end,
		ProjectID: uuid.New(),
	}
	activities := []*Activity{activity}

	project := &Project{
		ID:    activity.ProjectID,
		Title: "My Project",
	}
	projects := []*Project{project}

	var buffer bytes.Buffer

	err := a.WriteAsCSV(activities, projects, &buffer)

	is.NoErr(err)
	csv := buffer.String()

	is.True(strings.Contains(csv, "Date"))
	is.True(strings.Contains(csv, "My Project"))
	is.True(strings.Contains(csv, "11:00"))
	is.True(strings.Contains(csv, "11:30"))
}

func TestWriteAsExcel(t *testing.T) {
	is := is.New(t)

	a := &app{}

	start, _ := time.Parse(time.RFC3339, "2021-11-12T11:00:00.000Z")
	end, _ := time.Parse(time.RFC3339, "2021-11-12T11:30:00.000Z")

	activity := &Activity{
		Start:     start,
		End:       end,
		ProjectID: uuid.New(),
	}
	activities := []*Activity{activity}

	project := &Project{
		ID:    activity.ProjectID,
		Title: "My Project",
	}
	projects := []*Project{project}

	var buffer bytes.Buffer

	err := a.WriteAsExcel(activities, projects, &buffer)

	is.NoErr(err)
}
