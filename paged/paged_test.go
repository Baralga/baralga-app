package paged

import (
	"net/url"
	"testing"

	"github.com/matryer/is"
)

func TestPageOfTotalEven(t *testing.T) {
	// Arrange
	is := is.New(t)

	pageParams := &PageParams{
		Page: 1,
		Size: 10,
	}

	// Act
	page := pageParams.PageOfTotal(100)

	// Assert
	is.Equal(page.TotalPages, 10)
}

func TestPageOfTotalUneven(t *testing.T) {
	// Arrange
	is := is.New(t)

	pageParams := &PageParams{
		Page: 1,
		Size: 10,
	}

	// Act
	page := pageParams.PageOfTotal(11)

	// Assert
	is.Equal(page.TotalPages, 2)
}

func TestPageParamsFromQueryWithoutParams(t *testing.T) {
	// Arrange
	is := is.New(t)
	queryParams := make(url.Values)

	// Act
	pageParams := PageParamsFromQuery(queryParams, 50)

	// Assert
	is.Equal(pageParams.Page, 0)
	is.Equal(pageParams.Size, 50)
}

func TestPageParamsFromQueryWitParams(t *testing.T) {
	// Arrange
	is := is.New(t)
	queryParams := make(url.Values)
	queryParams["p"] = []string{"1"}

	// Act
	pageParams := PageParamsFromQuery(queryParams, 50)

	// Assert
	is.Equal(pageParams.Page, 1)
	is.Equal(pageParams.Size, 50)
}
