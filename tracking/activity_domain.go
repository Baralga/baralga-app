package tracking

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/baralga/shared/paged"
	time_utils "github.com/baralga/tracking/time"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	SortOrderAsc  string = "asc"
	SortOrderDesc string = "desc"
)

var ErrActivityNotFound = errors.New("activity not found")

// Activity represents a tracked time for a project
type Activity struct {
	ID             uuid.UUID
	Start          time.Time
	End            time.Time
	Description    string
	ProjectID      uuid.UUID
	OrganizationID uuid.UUID
	Username       string
	Tags           []*Tag // slice of Tag objects with full information
}

// Tag represents a tag that can be associated with activities
type Tag struct {
	ID             uuid.UUID
	Name           string // normalized (lowercase)
	Color          string // hex color code
	OrganizationID uuid.UUID
	CreatedAt      time.Time
}

// TagReportItem represents a single tag's time report data
type TagReportItem struct {
	TagName                string
	TagColor               string
	Year                   int
	Quarter                int
	Month                  int
	Week                   int
	Day                    int
	DurationInMinutesTotal int
	ActivityCount          int
}

// DurationFormatted is the tag report duration as formatted string (e.g. 1:15 h)
func (t *TagReportItem) DurationFormatted() string {
	return time_utils.FormatMinutesAsDuration(float64(t.DurationInMinutesTotal))
}

// ActivityFilter reprensents a filter for activities
type ActivityFilter struct {
	Timespan  string
	sortBy    string
	sortOrder string
	start     time.Time
	end       time.Time
	tags      []string // tag names to filter by
}

type ActivityTimeReportItem struct {
	Year                   int
	Quarter                int
	Month                  int
	Week                   int
	Day                    int
	DurationInMinutesTotal int
}

type ActivitiesPaged struct {
	Activities []*Activity
	Page       *paged.Page
}

type ActivitiesFilter struct {
	Start          time.Time
	End            time.Time
	SortBy         string
	SortOrder      string
	Username       string
	OrganizationID uuid.UUID
}

func IsValidActivitySortField(f string) bool {
	switch strings.ToLower(f) {
	case "project":
		return true
	case "start":
		return true
	default:
		return false
	}
}

func IsValidSortOrder(f string) bool {
	switch strings.ToLower(f) {
	case SortOrderAsc:
		return true
	case SortOrderDesc:
		return true
	default:
		return false
	}
}

type ActivityRepository interface {
	TimeReportByDay(ctx context.Context, filter *ActivitiesFilter) ([]*ActivityTimeReportItem, error)
	TimeReportByWeek(ctx context.Context, filter *ActivitiesFilter) ([]*ActivityTimeReportItem, error)
	TimeReportByMonth(ctx context.Context, filter *ActivitiesFilter) ([]*ActivityTimeReportItem, error)
	TimeReportByQuarter(ctx context.Context, filter *ActivitiesFilter) ([]*ActivityTimeReportItem, error)
	ProjectReport(ctx context.Context, filter *ActivitiesFilter) ([]*ActivityProjectReportItem, error)
	FindActivities(ctx context.Context, filter *ActivitiesFilter, pageParams *paged.PageParams) (*ActivitiesPaged, []*Project, error)
	InsertActivity(ctx context.Context, activity *Activity) (*Activity, error)
	FindActivityByID(ctx context.Context, activityID uuid.UUID, organizationID uuid.UUID) (*Activity, error)
	DeleteActivityByID(ctx context.Context, organizationID, activityID uuid.UUID) error
	DeleteActivityByIDAndUsername(ctx context.Context, organizationID, activityID uuid.UUID, username string) error
	UpdateActivity(ctx context.Context, organizationID uuid.UUID, activity *Activity) (*Activity, error)
	UpdateActivityByUsername(ctx context.Context, organizationID uuid.UUID, activity *Activity, username string) (*Activity, error)
}

// TagRepository manages tag CRUD operations
type TagRepository interface {
	// FindTagsByOrganization returns all tags for a specific organization for autocomplete
	FindTagsByOrganization(ctx context.Context, organizationID uuid.UUID, query string) ([]*Tag, error)
	// FindOrCreateTag gets existing or creates new tag for organization
	FindOrCreateTag(ctx context.Context, name string, organizationID uuid.UUID) (*Tag, error)
	// SyncTagsForActivity creates/updates tag relationships when activity is saved
	SyncTagsForActivity(ctx context.Context, activityID uuid.UUID, organizationID uuid.UUID, tags []*Tag) error
	// DeleteUnusedTags cleanup method for organization-level cleanup
	DeleteUnusedTags(ctx context.Context, organizationID uuid.UUID) error
	// GetTagReportData retrieves tag report data with time aggregation
	GetTagReportData(ctx context.Context, filter *ActivitiesFilter, aggregateBy string) ([]*TagReportItem, error)
}

// DurationFormatted is the activity duration as formatted string (e.g. 1:15 h)
func (i *ActivityTimeReportItem) DurationFormatted() string {
	return time_utils.FormatMinutesAsDuration(float64(i.DurationInMinutesTotal))
}

type ActivityProjectReportItem struct {
	ProjectID              uuid.UUID
	ProjectTitle           string
	DurationInMinutesTotal int
}

// DurationFormatted is the activity duration as formatted string (e.g. 1:15 h)
func (i *ActivityProjectReportItem) DurationFormatted() string {
	return time_utils.FormatMinutesAsDuration(float64(i.DurationInMinutesTotal))
}

// AsTime returns the report item as time.Time
func (i *ActivityTimeReportItem) AsTime() time.Time {
	t, _ := time.Parse("2006-1-2", fmt.Sprintf("%v-%v-%v", i.Year, i.Month, i.Day))
	return t
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

// Tags returns the filter's tag names
func (f *ActivityFilter) Tags() []string {
	return f.tags
}

// WithTags returns a new filter with the specified tags
func (f *ActivityFilter) WithTags(tags []string) *ActivityFilter {
	return &ActivityFilter{
		Timespan:  f.Timespan,
		sortBy:    f.sortBy,
		sortOrder: f.sortOrder,
		start:     f.start,
		end:       f.end,
		tags:      tags,
	}
}

func (f *ActivityFilter) Home() *ActivityFilter {
	return &ActivityFilter{
		Timespan: f.Timespan,
		start:    time.Now(),
		tags:     f.tags,
	}
}

func (f *ActivityFilter) Next() *ActivityFilter {
	nextFilter := &ActivityFilter{
		Timespan: f.Timespan,
		start:    f.start,
		end:      f.end,
		tags:     f.tags,
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
		tags:     f.tags,
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

func (f *ActivityFilter) WithSortToggle(sortBy string) *ActivityFilter {
	filterWithSort := &ActivityFilter{
		Timespan: f.Timespan,
		sortBy:   sortBy,
		start:    f.start,
		end:      f.end,
		tags:     f.tags,
	}

	if f.sortOrder == "desc" {
		filterWithSort.sortOrder = "asc"
	} else {
		filterWithSort.sortOrder = "desc"
	}

	return filterWithSort
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
		q := time_utils.Quarter(f.Start())
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
			time_utils.FormatDateDEShort(f.Start()),
			time_utils.FormatDateDEShort(f.End()),
		)
	case TimespanDay:
		return time_utils.FormatDateDEShort(f.Start())
	default:
		return fmt.Sprintf(
			"%v - %v",
			time_utils.FormatDateDEShort(f.Start()),
			time_utils.FormatDateDEShort(f.End().AddDate(0, 0, -1)),
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
func (a *Activity) DurationHours() int {
	return int(a.duration().Hours())
}

// DurationMinutes is the activity duration in minutes of unfinished hour (e.g. 15)
func (a *Activity) DurationMinutes() int {
	m := int(a.duration().Minutes())
	return m % 60
}

// DurationMinutesTotal is the activity duration in minutes in total (unrounded)
func (a *Activity) DurationMinutesTotal() int {
	return int(a.duration().Minutes())
}

// DurationDecimal is the activity duration as decimal (e.g. 0.75)
func (a *Activity) DurationDecimal() float64 {
	return a.duration().Minutes() / 60.0
}

// DurationFormatted is the activity duration as formatted string (e.g. 1:15 h)
func (a *Activity) DurationFormatted() string {
	return time_utils.FormatMinutesAsDuration(float64(a.DurationMinutesTotal()))
}

func (a *Activity) duration() time.Duration {
	return a.End.Sub(a.Start)
}
