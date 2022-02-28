package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestHandleReportPage(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ProjectRepository:  NewInMemProjectRepository(),
		ActivityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("GET", "/reports", nil)
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleReportPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "Report Activities # Baralga"))
}

func TestHandleReportPageWithTimeByDay(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ProjectRepository:  NewInMemProjectRepository(),
		ActivityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("GET", "/reports?c=time:d", nil)
	r.Header.Add("HX-Request", "true")
	r.Header.Add("HX-Target", "baralga__report_content")
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleReportPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "id=\"time-report-by-day\""))
}

func TestHandleReportPageWithTimeByWeek(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ProjectRepository:  NewInMemProjectRepository(),
		ActivityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("GET", "/reports?c=time:w&t=year", nil)
	r.Header.Add("HX-Request", "true")
	r.Header.Add("HX-Target", "baralga__report_content")
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleReportPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "id=\"time-report-by-week\""))
}

func TestHandleReportPageWithTimeByMonth(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ProjectRepository:  NewInMemProjectRepository(),
		ActivityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("GET", "/reports?c=time:m&t=year", nil)
	r.Header.Add("HX-Request", "true")
	r.Header.Add("HX-Target", "baralga__report_content")
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleReportPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "id=\"time-report-by-month\""))
}

func TestHandleReportPageWithTimeByQuarter(t *testing.T) {
	is := is.New(t)
	httpRec := httptest.NewRecorder()

	a := &app{
		Config:             &config{},
		ProjectRepository:  NewInMemProjectRepository(),
		ActivityRepository: NewInMemActivityRepository(),
	}

	r, _ := http.NewRequest("GET", "/reports?c=time:q&t=year", nil)
	r.Header.Add("HX-Request", "true")
	r.Header.Add("HX-Target", "baralga__report_content")
	r = r.WithContext(context.WithValue(r.Context(), contextKeyPrincipal, &Principal{}))

	a.HandleReportPage()(httpRec, r)
	is.Equal(httpRec.Result().StatusCode, http.StatusOK)

	htmlBody := httpRec.Body.String()
	is.True(strings.Contains(htmlBody, "id=\"time-report-by-quarter\""))
}

func TestReportViewFromQueryParams(t *testing.T) {
	is := is.New(t)

	t.Run("view without params", func(t *testing.T) {
		// Arrange
		params := make(url.Values)

		// Act
		view := reportViewFromQueryParams(params, "")

		// Assert
		is.Equal(view.main, "general")
	})

	t.Run("view with week and month view", func(t *testing.T) {
		// Arrange
		params := make(url.Values)
		params["t"] = []string{"week"}
		params["c"] = []string{"time:m"}

		// Act
		view := reportViewFromQueryParams(params, "week")

		// Assert
		is.Equal(view.main, "time")
		is.Equal(view.sub, "w")
	})

	t.Run("view with week and quarter view", func(t *testing.T) {
		// Arrange
		params := make(url.Values)
		params["t"] = []string{"week"}
		params["c"] = []string{"time:q"}

		// Act
		view := reportViewFromQueryParams(params, "week")

		// Assert
		is.Equal(view.main, "time")
		is.Equal(view.sub, "w")
	})

	t.Run("view with week and day view", func(t *testing.T) {
		// Arrange
		params := make(url.Values)
		params["t"] = []string{"week"}
		params["c"] = []string{"time:d"}

		// Act
		view := reportViewFromQueryParams(params, "week")

		// Assert
		is.Equal(view.main, "time")
		is.Equal(view.sub, "d")
	})

	t.Run("view with month and quarter view", func(t *testing.T) {
		// Arrange
		params := make(url.Values)
		params["t"] = []string{"month"}
		params["c"] = []string{"time:q"}

		// Act
		view := reportViewFromQueryParams(params, "month")

		// Assert
		is.Equal(view.main, "time")
		is.Equal(view.sub, "m")
	})

	t.Run("view with day and quarter view", func(t *testing.T) {
		// Arrange
		params := make(url.Values)
		params["t"] = []string{"day"}
		params["c"] = []string{"time:q"}

		// Act
		view := reportViewFromQueryParams(params, "day")

		// Assert
		is.Equal(view.main, "time")
		is.Equal(view.sub, "d")
	})

	t.Run("view with day and month view", func(t *testing.T) {
		// Arrange
		params := make(url.Values)
		params["t"] = []string{"day"}
		params["c"] = []string{"time:m"}

		// Act
		view := reportViewFromQueryParams(params, "day")

		// Assert
		is.Equal(view.main, "time")
		is.Equal(view.sub, "d")
	})

	t.Run("view with day and week view", func(t *testing.T) {
		// Arrange
		params := make(url.Values)
		params["t"] = []string{"day"}
		params["c"] = []string{"time:w"}

		// Act
		view := reportViewFromQueryParams(params, "day")

		// Assert
		is.Equal(view.main, "time")
		is.Equal(view.sub, "d")
	})
}
