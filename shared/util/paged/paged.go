package paged

import (
	"math"
	"net/http"
	"net/url"
	"strconv"
)

type Page struct {
	Size          int `json:"size"`
	TotalElements int `json:"totalElements"`
	TotalPages    int `json:"totalPages"`
	Number        int `json:"number"`
}

type PageParams struct {
	Page int
	Size int
}

// PageParamsFromQuery read the paging parameters from the url query params
func PageParamsFromQuery(queryParams url.Values, size int) PageParams {
	page := 0

	// read page from query params
	if len(queryParams["p"]) > 0 {
		p, err := strconv.Atoi(queryParams["p"][0])
		if err == nil {
			page = p
		}
	}

	return PageParams{
		Size: size,
		Page: page,
	}
}

func (p *PageParams) Offset() int {
	return p.Page * p.Size
}

func (p *PageParams) PageOfTotal(total int) *Page {
	return &Page{
		Size:          p.Size,
		Number:        p.Page,
		TotalElements: total,
		TotalPages:    int(math.Ceil(float64(total) / float64(p.Size))),
	}
}

func PageParamsOf(r *http.Request) *PageParams {
	pageParams := &PageParams{
		Page: 0,
		Size: 50,
	}

	pageQueryParam := r.URL.Query().Get("page")
	if pageQueryParam != "" {
		page, err := strconv.Atoi(pageQueryParam)
		if err == nil {
			pageParams.Page = page
		}
	}

	sizeQueryParam := r.URL.Query().Get("size")
	if pageQueryParam != "" {
		size, err := strconv.Atoi(sizeQueryParam)
		if err == nil {
			pageParams.Size = size
		}
	}

	return pageParams
}
