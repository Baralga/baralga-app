package tracking

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/baralga/shared"
	"github.com/baralga/shared/hal"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

// Helper function to create a properly initialized ActivityService for tests
func createTestActivityServiceForRest(repo ActivityRepository) *ActitivityService {
	tagRepo := NewInMemTagRepository()
	tagService := NewTagService(tagRepo)
	return &ActitivityService{
		repositoryTxer:     shared.NewInMemRepositoryTxer(),
		activityRepository: repo,
		tagRepository:      tagRepo,
		tagService:         tagService,
	}
}

func TestMapToActivity(t *testing.T) {
	is := is.New(t)

	activityModel := &activityModel{
		ID:    "d9fbfab6-2750-4703-8a7b-77498756d64a",
		Start: "2021-11-06T21:37:00",
		End:   "2021-11-06T21:37:00",
		Links: hal.NewLinks(
			hal.NewLink("project", "/api/projects/efa45cae-5dc7-412a-887f-945ddbb0a23f"),
		),
	}

	activity, err := mapToActivity(activityModel)

	is.NoErr(err)
	is.Equal(activityModel.ID, activity.ID.String())
	is.Equal(uuid.MustParse("efa45cae-5dc7-412a-887f-945ddbb0a23f"), activity.ProjectID)
	is.Equal(2021, activity.Start.Year())
	is.Equal(2021, activity.End.Year())
}

func TestMapToActivityIdNotValid(t *testing.T) {
	is := is.New(t)

	activityModel := &activityModel{
		ID:    "no-uuid",
		Start: "2021-11-06T21:37:00",
		End:   "2021-11-06T21:37:00",
		Links: hal.NewLinks(
			hal.NewLink("project", "/api/projects/efa45cae-5dc7-412a-887f-945ddbb0a23f"),
		),
	}

	_, err := mapToActivity(activityModel)

	is.True(err != nil)
}

func TestMapToActivityModel(t *testing.T) {
	is := is.New(t)

	start, _ := time.Parse(time.RFC3339, "2021-11-12T11:00:00.000Z")
	end, _ := time.Parse(time.RFC3339, "2021-11-12T11:30:00.000Z")

	activity := &Activity{
		ID:        uuid.New(),
		Start:     start,
		End:       end,
		ProjectID: uuid.New(),
	}

	activityModel := mapToActivityModel(activity)

	is.Equal(activity.ID.String(), activityModel.ID)
	is.True(strings.Contains(activityModel.Links.HrefOf("project"), activity.ProjectID.String()))
	is.True(strings.Contains(activityModel.Start, "2021"))
	is.True(strings.Contains(activityModel.End, "2021"))
}

func TestHandleGetActivity(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("GET", "/api/activities/00000000-0000-0000-2222-000000000001", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleGetActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
}

func TestHandleGetActivityNotFound(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("GET", "/api/activities/d9fbfab6-2750-4703-8a7b-77498756d64a", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "d9fbfab6-2750-4703-8a7b-77498756d64a")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleGetActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotFound)
}

func TestHandleGetActivityIdNotValid(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("GET", "/api/activities/not-a-uuid", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "not-a-uuid")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleGetActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusBadRequest)
}

func TestHandleGetActivitiesWithUrlParams(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	activityRepository := NewInMemActivityRepository()
	a := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: activityRepository,
		actitivityService: &ActitivityService{
			activityRepository: activityRepository,
		},
	}

	r, _ := http.NewRequest("GET", "/api/activities?start=2021-10-01&end=2022-10-01", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleGetActivities()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	activitiesModel := &activitiesModel{}
	err := json.NewDecoder(httpRec.Body).Decode(activitiesModel)
	is.NoErr(err)
	is.Equal(1, len(activitiesModel.ActivityModels))
}

func TestHandleGetActivitiesWithTimespanUrlParams(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	activityRepository := NewInMemActivityRepository()
	a := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: activityRepository,
		actitivityService: &ActitivityService{
			activityRepository: activityRepository,
		},
	}

	r, _ := http.NewRequest("GET", "/api/activities?t=week&v=2020-3", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleGetActivities()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	activitiesModel := &activitiesModel{}
	err := json.NewDecoder(httpRec.Body).Decode(activitiesModel)
	is.NoErr(err)
	is.Equal(1, len(activitiesModel.ActivityModels))
}

func TestHandleGetActivitiesWithTimespanUrlParamsAsCSV(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	activityRepository := NewInMemActivityRepository()
	a := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: activityRepository,
		actitivityService: &ActitivityService{
			activityRepository: activityRepository,
		},
	}

	r, _ := http.NewRequest("GET", "/api/activities?t=week&v=2020-3", nil)
	r.Header.Set("Content-Type", "text/csv")
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleGetActivities()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	csv := httpRec.Body.String()
	is.True(strings.Contains(csv, "Date"))
}

func TestHandleGetActivitiesWithTimespanUrlParamsAsExcel(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()

	c := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: repo,
		actitivityService: &ActitivityService{
			activityRepository: repo,
		},
	}

	r, _ := http.NewRequest("GET", "/api/activities?t=week&v=2020-3", nil)
	r.Header.Set("Content-Type", "application/vnd.ms-excel")
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	c.HandleGetActivities()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
}

func TestHandleCreateActivity(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()
	config := &shared.Config{}

	c := &ActivityRestHandlers{
		config:             config,
		activityRepository: repo,
		actitivityService:  createTestActivityServiceForRest(repo),
	}

	countBefore := len(repo.activities)
	body := `
	{
		"id":null,
		"start":"2021-11-06T21:37:00",
		"end":"2021-11-06T21:37:00",
		"description":"",
		"_links":{
		   "project":{
			  "href":"http://localhost:8080/api/projects/f4b1087c-8fbb-4c8d-bbb7-ab4d46da16ea"
		   }
		}
	 }
	`

	r, _ := http.NewRequest("POST", "/api/activities", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	c.HandleCreateActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusCreated)
	is.Equal(countBefore+1, len(repo.activities))
}

func TestHandleCreateInvalidActivity(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()
	a := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: repo,
	}

	body := `
	{
		"id":null,
		"start":"2021-11-06T21:37:00",
		"end":"2021-11-06T21:37:00",
		"description":"Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu, consequat vitae, eleifend ac, enim. Aliquam lorem ante, dapibus in, viverra quis, feugiat a, tellus. Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue. Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque sed ipsum. Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id, lorem. Maecenas nec odio et ante tincidunt tempus. Donec vitae sapien ut libero venenatis faucibus. Nullam quis ante. Etiam sit amet orci eget eros faucibus tincidunt. Duis leo. Sed fringilla mauris sit amet nibh. Donec sodales sagittis magna. Sed consequat, leo eget bibendum sodales, augue velit cursus nunc, quis gravida magna mi a libero. Fusce vulputate eleifend sapien. Vestibulum purus quam, scelerisque ut, mollis sed, nonummy id, metus. Nullam accumsan lorem in dui. Cras ultricies mi eu turpis hendrerit fringilla. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; In ac dui quis mi consectetuer lacinia. Nam pretium turpis et arcu. Duis arcu tortor, suscipit eget, imperdiet nec, imperdiet iaculis, ipsum. Sed aliquam ultrices mauris. Integer ante arcu, accumsan a, consectetuer eget, posuere ut, mauris. Praesent adipiscing. Phasellus ullamcorper ipsum rutrum nunc. Nunc nonummy metus. Vestibulum volutpat pretium libero. Cras id dui. Aenean ut eros et nisl sagittis vestibulum. Nullam nulla eros, ultricies sit amet, nonummy id, imperdiet feugiat, pede. Sed lectus. Donec mollis hendrerit risus. Phasellus nec sem in justo pellentesque facilisis. Etiam imperdiet imperdiet orci. Nunc nec neque. Phasellus leo dolor, tempus non, auctor et, hendrerit quis, nisi. Curabitur ligula sapien, tincidunt non, euismod vitae, posuere imperdiet, leo. Maecenas malesuada. Praesent congue erat at massa. Sed cursus turpis vitae tortor. Donec posuere vulputate arcu. Phasellus accumsan cursus velit. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Sed aliquam, nisi quis porttitor congue, elit erat euismod orci, ac placerat dolor lectus quis orci. Phasellus consectetuer vestibulum elit. Aenean tellus metus, bibendum sed, posuere ac, mattis non, nunc. Vestibulum fringilla pede sit amet augue. In turpis. Pellentesque posuere. Praesent turpis. Aenean posuere, tortor sed cursus feugiat, nunc augue blandit nunc, eu sollicitudin urna dolor sagittis lacus. Donec elit libero, sodales nec, volutpat a, suscipit non, turpis. Nullam sagittis. Suspendisse pulvinar, augue ac venenatis condimentum, sem libero volutpat nibh, nec pellentesque velit pede quis nunc. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Fusce id purus. Ut varius tincidunt libero. Phasellus dolor. Maecenas vestibulum mollis diam. ",
		"_links":{
		   "project":{
			  "href":"http://localhost:8080/api/projects/f4b1087c-8fbb-4c8d-bbb7-ab4d46da16ea"
		   }
		}
	 }
	`

	r, _ := http.NewRequest("POST", "/api/activities", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleCreateActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusBadRequest)
}

func TestHandleDeleteActivityAsAdmin(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()
	config := &shared.Config{}

	c := &ActivityRestHandlers{
		config:             config,
		activityRepository: repo,
		actitivityService:  createTestActivityServiceForRest(repo),
	}

	r, _ := http.NewRequest("DELETE", "/api/activities/00000000-0000-0000-2222-000000000001", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Username: "admin",
		Roles:    []string{"ROLE_ADMIN"},
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	c.HandleDeleteActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
	is.Equal(0, len(repo.activities))
}

func TestHandleDeleteActivityAsMatchingUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()
	config := &shared.Config{}

	c := &ActivityRestHandlers{
		config:             config,
		activityRepository: repo,
		actitivityService:  createTestActivityServiceForRest(repo),
	}

	r, _ := http.NewRequest("DELETE", "/api/activities/00000000-0000-0000-2222-000000000001", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Username: "user1",
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	c.HandleDeleteActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)
	is.Equal(0, len(repo.activities))
}

func TestHandleDeleteActivityIdNotValid(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("DELETE", "/api/activities/not-a-uuid", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "not-a-uuid")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleDeleteActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusBadRequest)
}

func TestHandleCreateActivityWithInvalidBody(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: NewInMemActivityRepository(),
	}

	body := `
	{
		INVALID!!
	 }
	`

	r, _ := http.NewRequest("POST", "/api/activities", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	a.HandleCreateActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusBadRequest)
}

func TestHandleDeleteActivityAsNonMatchingUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()

	c := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: repo,
		actitivityService:  createTestActivityServiceForRest(repo),
	}

	r, _ := http.NewRequest("DELETE", "/api/activities/00000000-0000-0000-2222-000000000001", nil)
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Username: "otherUser",
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	c.HandleDeleteActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotFound)
	is.Equal(1, len(repo.activities))
}

func TestHandleUpdateActivity(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()

	c := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: repo,
		actitivityService:  createTestActivityServiceForRest(repo),
	}

	body := `
	{
		"id":null,
		"start":"2021-11-06T21:37:00",
		"end":"2021-11-06T21:37:00",
		"description": "My updated Description",
		"_links":{
		   "project":{
			  "href":"http://localhost:8080/api/projects/f4b1087c-8fbb-4c8d-bbb7-ab4d46da16ea"
		   }
		}
	 }
	`

	r, _ := http.NewRequest("POST", "/api/activities", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	c.HandleUpdateActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	activityUpdate, err := repo.FindActivityByID(context.Background(), uuid.MustParse("00000000-0000-0000-2222-000000000001"), shared.OrganizationIDSample)
	is.NoErr(err)
	is.Equal("My updated Description", activityUpdate.Description)
}

func TestHandleUpdateInvalidActivity(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()
	c := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: repo,
	}

	body := `
	{
		"id":null,
		"start":"2021-11-06T21:37:00",
		"end":"2021-11-06T21:37:00",
		"description": "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu, consequat vitae, eleifend ac, enim. Aliquam lorem ante, dapibus in, viverra quis, feugiat a, tellus. Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue. Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque sed ipsum. Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id, lorem. Maecenas nec odio et ante tincidunt tempus. Donec vitae sapien ut libero venenatis faucibus. Nullam quis ante. Etiam sit amet orci eget eros faucibus tincidunt. Duis leo. Sed fringilla mauris sit amet nibh. Donec sodales sagittis magna. Sed consequat, leo eget bibendum sodales, augue velit cursus nunc, quis gravida magna mi a libero. Fusce vulputate eleifend sapien. Vestibulum purus quam, scelerisque ut, mollis sed, nonummy id, metus. Nullam accumsan lorem in dui. Cras ultricies mi eu turpis hendrerit fringilla. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; In ac dui quis mi consectetuer lacinia. Nam pretium turpis et arcu. Duis arcu tortor, suscipit eget, imperdiet nec, imperdiet iaculis, ipsum. Sed aliquam ultrices mauris. Integer ante arcu, accumsan a, consectetuer eget, posuere ut, mauris. Praesent adipiscing. Phasellus ullamcorper ipsum rutrum nunc. Nunc nonummy metus. Vestibulum volutpat pretium libero. Cras id dui. Aenean ut eros et nisl sagittis vestibulum. Nullam nulla eros, ultricies sit amet, nonummy id, imperdiet feugiat, pede. Sed lectus. Donec mollis hendrerit risus. Phasellus nec sem in justo pellentesque facilisis. Etiam imperdiet imperdiet orci. Nunc nec neque. Phasellus leo dolor, tempus non, auctor et, hendrerit quis, nisi. Curabitur ligula sapien, tincidunt non, euismod vitae, posuere imperdiet, leo. Maecenas malesuada. Praesent congue erat at massa. Sed cursus turpis vitae tortor. Donec posuere vulputate arcu. Phasellus accumsan cursus velit. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Sed aliquam, nisi quis porttitor congue, elit erat euismod orci, ac placerat dolor lectus quis orci. Phasellus consectetuer vestibulum elit. Aenean tellus metus, bibendum sed, posuere ac, mattis non, nunc. Vestibulum fringilla pede sit amet augue. In turpis. Pellentesque posuere. Praesent turpis. Aenean posuere, tortor sed cursus feugiat, nunc augue blandit nunc, eu sollicitudin urna dolor sagittis lacus. Donec elit libero, sodales nec, volutpat a, suscipit non, turpis. Nullam sagittis. Suspendisse pulvinar, augue ac venenatis condimentum, sem libero volutpat nibh, nec pellentesque velit pede quis nunc. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; Fusce id purus. Ut varius tincidunt libero. Phasellus dolor. Maecenas vestibulum mollis diam. ",
		"_links":{
		   "project":{
			  "href":"http://localhost:8080/api/projects/f4b1087c-8fbb-4c8d-bbb7-ab4d46da16ea"
		   }
		}
	 }
	`

	r, _ := http.NewRequest("POST", "/api/activities", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Roles: []string{"ROLE_ADMIN"},
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	c.HandleUpdateActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusBadRequest)
}

func TestHandleUpdateActivityAsUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()

	c := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: repo,
		actitivityService:  createTestActivityServiceForRest(repo),
	}

	body := `
	{
		"id":null,
		"start":"2021-11-06T21:37:00",
		"end":"2021-11-06T21:37:00",
		"description": "My updated Description",
		"_links":{
		   "project":{
			  "href":"http://localhost:8080/api/projects/f4b1087c-8fbb-4c8d-bbb7-ab4d46da16ea"
		   }
		}
	 }
	`

	r, _ := http.NewRequest("POST", "/api/activities", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Username: "user1",
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	c.HandleUpdateActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	activityUpdate, err := repo.FindActivityByID(context.Background(), uuid.MustParse("00000000-0000-0000-2222-000000000001"), shared.OrganizationIDSample)
	is.NoErr(err)
	is.Equal("My updated Description", activityUpdate.Description)
}

func TestHandleUpdateActivityWithNonMatchingUser(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	repo := NewInMemActivityRepository()

	c := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: repo,
		actitivityService:  createTestActivityServiceForRest(repo),
	}

	body := `
	{
		"id":null,
		"start":"2021-11-06T21:37:00",
		"end":"2021-11-06T21:37:00",
		"description": "My updated Description",
		"_links":{
		   "project":{
			  "href":"http://localhost:8080/api/projects/f4b1087c-8fbb-4c8d-bbb7-ab4d46da16ea"
		   }
		}
	 }
	`

	r, _ := http.NewRequest("POST", "/api/activities", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Username: "otherUser",
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	c.HandleUpdateActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusNotFound)
}

func TestHandleUpdateActivityWithInvalidBody(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	c := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: NewInMemActivityRepository(),
	}

	body := `
	{
		INVALID!!
	 }
	`

	r, _ := http.NewRequest("POST", "/api/activities", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{
		Username: "otherUser",
	}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "00000000-0000-0000-2222-000000000001")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	c.HandleUpdateActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusBadRequest)
}

func TestHandleUpdateActivityIdNotValid(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &ActivityRestHandlers{
		config:             &shared.Config{},
		activityRepository: NewInMemActivityRepository(),
	}

	body := `
	{
		"id":null,
		"start":"2021-11-06T21:37:00",
		"end":"2021-11-06T21:37:00",
		"description": "My updated Description",
		"_links":{
		   "project":{
			  "href":"http://localhost:8080/api/projects/f4b1087c-8fbb-4c8d-bbb7-ab4d46da16ea"
		   }
		}
	 }
	`

	r, _ := http.NewRequest("POST", "/api/activities/not-a-uuid", strings.NewReader(body))
	r = r.WithContext(shared.ToContextWithPrincipal(r.Context(), &shared.Principal{}))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("activity-id", "not-a-uuid")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	a.HandleUpdateActivity()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusBadRequest)
}

func TestFilterFromQueryParams(t *testing.T) {
	is := is.New(t)

	t.Run("year filter without value", func(t *testing.T) {
		params := make(url.Values)
		params.Add("t", "year")

		filter, err := filterFromQueryParams(params)

		is.NoErr(err)
		is.Equal(time.Now().Year(), filter.Start().Year())
	})

	t.Run("year filter without value and sort order", func(t *testing.T) {
		params := make(url.Values)
		params.Add("t", "year")
		params.Add("sort", "project:asc")

		filter, err := filterFromQueryParams(params)

		is.NoErr(err)
		is.Equal(time.Now().Year(), filter.Start().Year())
		is.Equal("project", filter.sortBy)
		is.Equal(SortOrderAsc, filter.sortOrder)
	})

	t.Run("year filter from query params", func(t *testing.T) {
		params := make(url.Values)
		params.Add("t", "year")
		params.Add("v", "2021")

		filter, err := filterFromQueryParams(params)

		is.NoErr(err)
		is.Equal(2021, filter.Start().Year())
		is.Equal(time.January, filter.Start().Month())
		is.Equal(1, filter.Start().Day())
		is.Equal(0, filter.Start().Hour())
	})

	t.Run("year filter from invalid query params", func(t *testing.T) {
		params := make(url.Values)
		params.Add("t", "year")
		params.Add("v", "XXXX")

		_, err := filterFromQueryParams(params)

		is.True(err != nil)
	})

	t.Run("quarter filter from query params", func(t *testing.T) {
		params := make(url.Values)
		params.Add("t", "quarter")
		params.Add("v", "2021-2")

		filter, err := filterFromQueryParams(params)

		is.NoErr(err)
		is.Equal(2021, filter.Start().Year())
		is.Equal(time.April, filter.Start().Month())
		is.Equal(0, filter.Start().Hour())
	})

	t.Run("quarter filter from invalid query params", func(t *testing.T) {
		params := make(url.Values)
		params.Add("t", "quarter")
		params.Add("v", "XXXX-9")

		_, err := filterFromQueryParams(params)

		is.True(err != nil)
	})

	t.Run("month filter from query params", func(t *testing.T) {
		params := make(url.Values)
		params.Add("t", "month")
		params.Add("v", "2021-11")

		filter, err := filterFromQueryParams(params)

		is.NoErr(err)
		is.Equal(2021, filter.Start().Year())
		is.Equal(time.November, filter.Start().Month())
		is.Equal(1, filter.Start().Day())
		is.Equal(0, filter.Start().Hour())
	})

	t.Run("month filter from invalid query params", func(t *testing.T) {
		params := make(url.Values)
		params.Add("t", "month")
		params.Add("v", "2020-99")

		_, err := filterFromQueryParams(params)

		is.True(err != nil)
	})

	t.Run("week filter from query params 2022", func(t *testing.T) {
		params := make(url.Values)
		params.Add("t", "week")
		params.Add("v", "2021-1")

		filter, err := filterFromQueryParams(params)

		is.NoErr(err)
		is.Equal(2021, filter.Start().Year())
		is.Equal(time.January, filter.Start().Month())
		is.Equal(4, filter.Start().Day())
		is.Equal(0, filter.Start().Hour())
	})

	t.Run("week filter from query params 2023", func(t *testing.T) {
		params := make(url.Values)
		params.Add("t", "week")
		params.Add("v", "2023-10")

		filter, err := filterFromQueryParams(params)

		is.NoErr(err)
		is.Equal(2023, filter.Start().Year())
		is.Equal(time.March, filter.Start().Month())
		is.Equal(6, filter.Start().Day())
		is.Equal(0, filter.Start().Hour())
	})

	t.Run("week filter from invalid query params", func(t *testing.T) {
		params := make(url.Values)
		params.Add("t", "week")
		params.Add("v", "2020-ccc")

		_, err := filterFromQueryParams(params)

		is.True(err != nil)
	})

	t.Run("month filter from query params", func(t *testing.T) {
		params := make(url.Values)
		params.Add("t", "month")
		params.Add("v", "2021-03")

		filter, err := filterFromQueryParams(params)

		is.NoErr(err)
		is.Equal(2021, filter.Start().Year())
		is.Equal(time.March, filter.Start().Month())
	})

	t.Run("day filter from query params", func(t *testing.T) {
		params := make(url.Values)
		params.Add("t", "day")
		params.Add("v", "2021-11-10")

		filter, err := filterFromQueryParams(params)

		is.NoErr(err)
		is.Equal(2021, filter.Start().Year())
		is.Equal(time.November, filter.Start().Month())
	})

}
