package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/baralga/hal"
	"github.com/baralga/paged"
	"github.com/baralga/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/snabb/isoweek"
	"schneider.vip/problem"
)

type activitiesModel struct {
	*EmbeddedActivities `json:"_embedded"`
	Links               *hal.Links `json:"_links"`
}

// EmbeddedActivities contains embedded activities and projects
type EmbeddedActivities struct {
	ActivityModels []*activityModel `json:"activities"`
	ProjectModels  []*projectModel  `json:"projects"`
}

type activityModel struct {
	ID          string         `json:"id"`
	Start       string         `json:"start" validate:"required"`
	End         string         `json:"end" validate:"required"`
	Description string         `json:"description" validate:"max=500"`
	Duration    *durationModel `json:"duration"`
	Links       *hal.Links     `json:"_links"`
}

type durationModel struct {
	Hours     int     `json:"hours"`
	Minutes   int     `json:"minutes"`
	Decimal   float64 `json:"decimal"`
	Formatted string  `json:"formatted"`
}

// HandleGetActivities reads activities
func (a *app) HandleGetActivities() http.HandlerFunc {
	isProduction := a.isProduction()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		principal := r.Context().Value(contextKeyPrincipal).(*Principal)
		pageParams := paged.PageParamsOf(r)

		filter, err := filterFromQueryParams(r.URL.Query())
		if err != nil {
			util.RenderProblemJSON(w, isProduction, err)
			return
		}

		activities, projects, err := a.ReadActivitiesWithProjects(r.Context(), principal, filter, pageParams)
		if err != nil {
			util.RenderProblemJSON(w, isProduction, err)
			return
		}

		if r.URL.Query().Get("contentType") == "text/csv" || r.Header.Get("Content-Type") == "text/csv" {
			w.Header().Set("Content-Type", "text/csv")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"Activities_%v.csv\"", filter.String()))
			err := a.WriteAsCSV(activities, projects, w)
			if err != nil {
				util.RenderProblemJSON(w, isProduction, err)
				return
			}
			return
		}

		activityModels := mapToActivityModels(activities)
		projectModels := mapToProjectModels(principal, projects)

		activitiesModel := &activitiesModel{
			EmbeddedActivities: &EmbeddedActivities{
				ProjectModels:  projectModels,
				ActivityModels: activityModels,
			},
			Links: hal.NewLinks(
				hal.NewSelfLink(r.RequestURI),
				hal.NewLink("create", "/api/activities"),
			),
		}

		util.RenderJSON(w, activitiesModel)
	}
}

// HandleGetActivities creates an activity
func (a *app) HandleCreateActivity() http.HandlerFunc {
	isProduction := a.isProduction()
	validator := validator.New()
	return func(w http.ResponseWriter, r *http.Request) {
		var activityModel activityModel
		err := json.NewDecoder(r.Body).Decode(&activityModel)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusBadRequest)
			return
		}

		err = validator.Struct(activityModel)
		if err != nil {
			http.Error(w, problem.New(problem.Title("activity not valid")).JSONString(), http.StatusBadRequest)
			return
		}

		activityToCreate, err := mapToActivity(&activityModel)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusBadRequest)
			return
		}

		principal := r.Context().Value(contextKeyPrincipal).(*Principal)

		activity, err := a.CreateActivity(r.Context(), principal, activityToCreate)
		if err != nil {
			util.RenderProblemJSON(w, isProduction, err)
			return
		}

		activityModelCreated := mapToActivityModel(activity)

		w.WriteHeader(http.StatusCreated)
		util.RenderJSON(w, activityModelCreated)
	}
}

// HandleGetActivity reads an activity
func (a *app) HandleGetActivity() http.HandlerFunc {
	isProduction := a.isProduction()
	return func(w http.ResponseWriter, r *http.Request) {
		activityIDParam := chi.URLParam(r, "activity-id")
		principal := r.Context().Value(contextKeyPrincipal).(*Principal)

		activityID, err := uuid.Parse(activityIDParam)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusBadRequest)
			return
		}

		activity, err := a.ActivityRepository.FindActivityByID(r.Context(), activityID, principal.OrganizationID)
		if errors.Is(err, ErrActivityNotFound) {
			http.Error(w, problem.New(problem.Title("activity not found")).JSONString(), http.StatusNotFound)
			return
		}
		if err != nil {
			util.RenderProblemJSON(w, isProduction, err)
			return
		}

		activityModel := mapToActivityModel(activity)
		util.RenderJSON(w, activityModel)
	}
}

// HandleDeleteActivity deletes an activity
func (a *app) HandleDeleteActivity() http.HandlerFunc {
	isProduction := a.isProduction()
	return func(w http.ResponseWriter, r *http.Request) {
		activityIDParam := chi.URLParam(r, "activity-id")
		principal := r.Context().Value(contextKeyPrincipal).(*Principal)

		activityID, err := uuid.Parse(activityIDParam)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusBadRequest)
			return
		}

		err = a.DeleteActivityByID(r.Context(), principal, activityID)
		if errors.Is(err, ErrActivityNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			util.RenderProblemJSON(w, isProduction, err)
			return
		}
	}
}

// HandleUpdateActivity updates an activity
func (a *app) HandleUpdateActivity() http.HandlerFunc {
	isProduction := a.isProduction()
	validator := validator.New()
	return func(w http.ResponseWriter, r *http.Request) {
		activityIDParam := chi.URLParam(r, "activity-id")
		principal := r.Context().Value(contextKeyPrincipal).(*Principal)

		var activityModel activityModel
		err := json.NewDecoder(r.Body).Decode(&activityModel)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusBadRequest)
			return
		}

		err = validator.Struct(activityModel)
		if err != nil {
			http.Error(w, problem.New(problem.Title("activity not valid")).JSONString(), http.StatusBadRequest)
			return
		}

		activity, err := mapToActivity(&activityModel)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusBadRequest)
			return
		}

		activityID, err := uuid.Parse(activityIDParam)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusBadRequest)
			return
		}
		activity.ID = activityID

		activityUpdate, err := a.UpdateActivity(r.Context(), principal, activity)
		if errors.Is(err, ErrActivityNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			util.RenderProblemJSON(w, isProduction, err)
			return
		}

		activityModelUpdate := mapToActivityModel(activityUpdate)
		util.RenderJSON(w, activityModelUpdate)
	}
}

func mapToActivity(activityModel *activityModel) (*Activity, error) {
	var activityID uuid.UUID

	if activityModel.ID != "" {
		aID, err := uuid.Parse(activityModel.ID)
		if err != nil {
			return nil, err
		}
		activityID = aID
	}

	start, err := util.ParseDateTime(activityModel.Start)
	if err != nil {
		return nil, err
	}

	end, err := util.ParseDateTime(activityModel.End)
	if err != nil {
		return nil, err
	}

	projectHref := activityModel.Links.HrefOf("project")
	projectID, err := uuid.Parse(projectHref[strings.LastIndex(projectHref, "/")+1:])
	if err != nil {
		return nil, err
	}

	activity := &Activity{
		ID:          activityID,
		Start:       *start,
		End:         *end,
		ProjectID:   projectID,
		Description: activityModel.Description,
	}

	return activity, nil
}

func mapToActivityModel(activity *Activity) *activityModel {
	return &activityModel{
		ID:          activity.ID.String(),
		Description: activity.Description,
		Start:       util.FormatDateTime(&activity.Start),
		End:         util.FormatDateTime(&activity.End),
		Links: hal.NewLinks(
			hal.NewSelfLink(fmt.Sprintf("/api/activities/%s", activity.ID)),
			hal.NewLink("delete", fmt.Sprintf("/api/activities/%s", activity.ID)),
			hal.NewLink("edit", fmt.Sprintf("/api/activities/%s", activity.ID)),
			hal.NewLink("project", fmt.Sprintf("/api/projects/%s", activity.ProjectID)),
		),
		Duration: &durationModel{
			Hours:     activity.DurationHours(),
			Minutes:   activity.DurationMinutes(),
			Decimal:   activity.DurationDecimal(),
			Formatted: activity.DurationFormatted(),
		},
	}
}

func mapToActivityModels(activities []*Activity) []*activityModel {
	activityModels := make([]*activityModel, len(activities))

	for i, activity := range activities {
		activityModel := mapToActivityModel(activity)
		activityModels[i] = activityModel
	}

	return activityModels
}

func mapToProjectModels(principal *Principal, projects []*Project) []*projectModel {
	activityModels := make([]*projectModel, len(projects))

	for i, project := range projects {
		projectModel := mapToProjectModel(principal, project)
		activityModels[i] = projectModel
	}

	return activityModels
}

func filterFromQueryParams(params url.Values) (*ActivityFilter, error) {
	var timespan string
	value := ""
	fmt.Println(params["t"])
	if len(params["t"]) == 0 {
		timespan = TimespanCustom
	} else {
		timespan = params.Get("t")
	}

	filter := &ActivityFilter{
		Timespan: timespan,
	}

	if timespan != TimespanCustom && len(params["v"]) != 0 {
		return nil, errors.New("missing timespan value")
	}
	value = params.Get("v")

	switch timespan {
	case TimespanYear:
		start, err := time.Parse("2006", value)
		if err != nil {
			return nil, err
		}
		filter.start = start
	case TimespanQuarter:
		if !strings.Contains(value, "-") {
			return nil, errors.New("invalid quarter")
		}
		valueParts := strings.Split(value, "-")
		start, err := time.Parse("2006", valueParts[0])
		if err != nil {
			return nil, err
		}

		d := 24 * time.Hour
		start.Truncate(d)

		startQuarterOfYear, err := strconv.Atoi(valueParts[1])
		if err != nil {
			return nil, errors.New("invalid quarter")
		}
		filter.start = start.AddDate(0, 3*(startQuarterOfYear-1), 0)
	case TimespanMonth:
		start, err := time.Parse("2006-01", value)
		if err != nil {
			return nil, err
		}
		filter.start = start
	case TimespanWeek:
		if !strings.Contains(value, "-") {
			return nil, errors.New("invalid week")
		}
		valueParts := strings.Split(value, "-")

		startYear, err := strconv.Atoi(valueParts[0])
		if err != nil {
			return nil, err
		}
		startWeekOfYear, err := strconv.Atoi(valueParts[1])
		if err != nil {
			return nil, err
		}

		filter.start = isoweek.StartTime(startYear, startWeekOfYear, time.UTC)
	case TimespanDay:
		start, err := time.Parse("2006-01-02", value)
		if err != nil {
			return nil, err
		}
		filter.start = start
	case TimespanCustom:
		startParamValue := params.Get("start")
		if startParamValue != "" {
			startParam, err := util.ParseDate(startParamValue)
			if err != nil {
				return nil, err
			}
			filter.start = *startParam
		}

		endParamValue := params.Get("end")
		if endParamValue != "" {
			endParam, err := util.ParseDate(endParamValue)
			if err != nil {
				return nil, err
			}
			filter.end = *endParam
		}
	default:
		return nil, errors.New("invalid activity filter")
	}

	return filter, nil
}
