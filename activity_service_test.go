package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matryer/is"
)

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
