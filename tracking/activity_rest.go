package tracking

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/baralga/shared"
	"github.com/baralga/shared/hal"
	"github.com/baralga/shared/paged"
	time_utils "github.com/baralga/tracking/time"
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

type ActivityRestHandlers struct {
	config             *shared.Config
	actitivityService  *ActitivityService
	activityRepository ActivityRepository
}

func NewActivityRestHandlers(config *shared.Config, actitivityService *ActitivityService, activityRepository ActivityRepository) *ActivityRestHandlers {
	return &ActivityRestHandlers{
		config:             config,
		actitivityService:  actitivityService,
		activityRepository: activityRepository,
	}
}

func (a *ActivityRestHandlers) RegisterOpen(r chi.Router) {
}

func (a *ActivityRestHandlers) RegisterProtected(r chi.Router) {
	r.Get("/activities", a.HandleGetActivities())
	r.Post("/activities", a.HandleCreateActivity())
	r.Get("/activities/{activity-id}", a.HandleGetActivity())
	r.Delete("/activities/{activity-id}", a.HandleDeleteActivity())
	r.Patch("/activities/{activity-id}", a.HandleUpdateActivity())
	r.Get("/tags/autocomplete", a.HandleGetTagsAutocomplete())
}

// HandleGetActivities reads activities
func (a *ActivityRestHandlers) HandleGetActivities() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	actitivityService := a.actitivityService
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		principal := shared.MustPrincipalFromContext(r.Context())
		pageParams := paged.PageParamsOf(r)

		filter, err := filterFromQueryParams(r.URL.Query())
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, errors.New("invalid query params"))
			return
		}

		activitiesPage, projects, err := actitivityService.ReadActivitiesWithProjects(r.Context(), principal, filter, pageParams)
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		if r.URL.Query().Get("contentType") == "text/csv" || r.Header.Get("Content-Type") == "text/csv" {
			w.Header().Set("Content-Type", "text/csv")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"Activities_%v.csv\"", filter.String()))
			err := actitivityService.WriteAsCSV(activitiesPage.Activities, projects, w)
			if err != nil {
				shared.RenderProblemJSON(w, isProduction, err)
				return
			}
			return
		} else if r.URL.Query().Get("contentType") == "application/vnd.ms-excel" || r.Header.Get("Content-Type") == "application/vnd.ms-excel" {
			w.Header().Set("Content-Type", "!!")
			w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"Activities_%v.xlsx\"", filter.String()))
			err := actitivityService.WriteAsExcel(activitiesPage.Activities, projects, w)
			if err != nil {
				shared.RenderProblemJSON(w, isProduction, err)
				return
			}
			return
		}

		activityModels := mapToActivityModels(activitiesPage.Activities)
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

		shared.RenderJSON(w, activitiesModel)
	}
}

// HandleGetActivities creates an activity
func (a *ActivityRestHandlers) HandleCreateActivity() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	validator := validator.New()
	actitivityService := a.actitivityService
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

		principal := shared.MustPrincipalFromContext(r.Context())

		activity, err := actitivityService.CreateActivity(r.Context(), principal, activityToCreate)
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		activityModelCreated := mapToActivityModel(activity)

		w.WriteHeader(http.StatusCreated)
		shared.RenderJSON(w, activityModelCreated)
	}
}

// HandleGetActivity reads an activity
func (a *ActivityRestHandlers) HandleGetActivity() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	activityRepository := a.activityRepository
	return func(w http.ResponseWriter, r *http.Request) {
		activityIDParam := chi.URLParam(r, "activity-id")
		principal := shared.MustPrincipalFromContext(r.Context())

		activityID, err := uuid.Parse(activityIDParam)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusBadRequest)
			return
		}

		activity, err := activityRepository.FindActivityByID(r.Context(), activityID, principal.OrganizationID)
		if errors.Is(err, ErrActivityNotFound) {
			http.Error(w, problem.New(problem.Title("activity not found")).JSONString(), http.StatusNotFound)
			return
		}
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		activityModel := mapToActivityModel(activity)
		shared.RenderJSON(w, activityModel)
	}
}

// HandleDeleteActivity deletes an activity
func (a *ActivityRestHandlers) HandleDeleteActivity() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	actitivityService := a.actitivityService
	return func(w http.ResponseWriter, r *http.Request) {
		activityIDParam := chi.URLParam(r, "activity-id")
		principal := shared.MustPrincipalFromContext(r.Context())

		activityID, err := uuid.Parse(activityIDParam)
		if err != nil {
			http.Error(w, problem.New(problem.Wrap(err)).JSONString(), http.StatusBadRequest)
			return
		}

		err = actitivityService.DeleteActivityByID(r.Context(), principal, activityID)
		if errors.Is(err, ErrActivityNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		w.Header().Set("HX-Trigger", "baralga__activities-changed")
	}
}

// HandleUpdateActivity updates an activity
func (a *ActivityRestHandlers) HandleUpdateActivity() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	validator := validator.New()
	actitivityService := a.actitivityService
	return func(w http.ResponseWriter, r *http.Request) {
		activityIDParam := chi.URLParam(r, "activity-id")
		principal := shared.MustPrincipalFromContext(r.Context())

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

		activityUpdate, err := actitivityService.UpdateActivity(r.Context(), principal, activity)
		if errors.Is(err, ErrActivityNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		activityModelUpdate := mapToActivityModel(activityUpdate)
		shared.RenderJSON(w, activityModelUpdate)
	}
}

// tagAutocompleteModel represents a tag suggestion for autocomplete
type tagAutocompleteModel struct {
	Name string `json:"name"`
}

// tagsAutocompleteResponse represents the response for tag autocomplete
type tagsAutocompleteResponse struct {
	Tags []*tagAutocompleteModel `json:"tags"`
}

// HandleGetTagsAutocomplete handles tag autocomplete requests
func (a *ActivityRestHandlers) HandleGetTagsAutocomplete() http.HandlerFunc {
	isProduction := a.config.IsProduction()
	actitivityService := a.actitivityService
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		principal := shared.MustPrincipalFromContext(r.Context())

		// Get query parameter
		query := r.URL.Query().Get("q")
		if query == "" {
			// Return empty response if no query provided
			response := &tagsAutocompleteResponse{
				Tags: []*tagAutocompleteModel{},
			}
			shared.RenderJSON(w, response)
			return
		}

		// Validate query length to prevent abuse
		if len(query) > 100 {
			http.Error(w, problem.New(problem.Title("query parameter too long")).JSONString(), http.StatusBadRequest)
			return
		}

		// Get matching tags from service
		tags, err := actitivityService.GetTagsForAutocomplete(r.Context(), principal, query)
		if err != nil {
			shared.RenderProblemJSON(w, isProduction, err)
			return
		}

		// Convert to response model
		tagModels := make([]*tagAutocompleteModel, len(tags))
		for i, tag := range tags {
			tagModels[i] = &tagAutocompleteModel{
				Name: tag.Name,
			}
		}

		response := &tagsAutocompleteResponse{
			Tags: tagModels,
		}

		shared.RenderJSON(w, response)
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

	start, err := time_utils.ParseDateTime(activityModel.Start)
	if err != nil {
		return nil, err
	}

	end, err := time_utils.ParseDateTime(activityModel.End)
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
		Start:       time_utils.FormatDateTime(activity.Start),
		End:         time_utils.FormatDateTime(activity.End),
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

func mapToProjectModels(principal *shared.Principal, projects []*Project) []*projectModel {
	activityModels := make([]*projectModel, len(projects))

	for i, project := range projects {
		projectModel := mapToProjectModel(principal, project)
		activityModels[i] = projectModel
	}

	return activityModels
}

func filterFromQueryParams(params url.Values) (*ActivityFilter, error) {
	if len(params["t"]) == 0 {
		params["t"] = []string{"week"}
	}

	var timespan string
	value := ""
	if len(params["t"]) == 0 {
		timespan = TimespanCustom
	} else {
		timespan = params["t"][0]
	}

	sortOrder := ""
	sortBy := ""
	if len(params["sort"]) != 0 {
		sortParam := strings.Split(params["sort"][0], ":")
		if len(sortParam) == 2 && IsValidActivitySortField(sortParam[0]) && IsValidSortOrder(sortParam[1]) {
			sortBy = sortParam[0]
			sortOrder = sortParam[1]
		}
	}

	filter := &ActivityFilter{
		Timespan:  timespan,
		sortBy:    sortBy,
		sortOrder: sortOrder,
	}

	if timespan == TimespanCustom && len(params["start"]) == 0 && len(params["end"]) == 0 {
		return nil, errors.New("missing timespan value")
	}

	if len(params["v"]) != 0 {
		value = params["v"][0]
	} else {
		value = filter.NewValue()
	}

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
		start = start.Truncate(d)

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
			startParam, err := time_utils.ParseDate(startParamValue)
			if err != nil {
				return nil, err
			}
			filter.start = *startParam
		}

		endParamValue := params.Get("end")
		if endParamValue != "" {
			endParam, err := time_utils.ParseDate(endParamValue)
			if err != nil {
				return nil, err
			}
			filter.end = *endParam
		}
	default:
		return nil, errors.New("invalid activity filter")
	}

	// Parse tag filters from URL parameters
	if tagParams := params["tags"]; len(tagParams) > 0 {
		var tags []string
		for _, tagParam := range tagParams {
			// Split comma-separated tags
			tagParts := strings.Split(tagParam, ",")
			for _, tag := range tagParts {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					tags = append(tags, strings.ToLower(tag))
				}
			}
		}
		filter.tags = tags
	}

	return filter, nil
}
