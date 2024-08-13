package domain

import (
	"fmt"
	"math"
	"slices"
	"strings"
)

type Filter struct {
	Page         int
	PageSize     int
	Sort         *string
	SafeSortList *[]string
}

func ValidateFilters(ev *ErrValidation, f *Filter) {
	ev.Evaluate(f.Page > 0, "page", "must be greater than zero")
	ev.Evaluate(f.Page <= 10_000_000, "page", "must be a max of 10 million")
	ev.Evaluate(f.PageSize > 0, "page_size", "must be greater than zero")
	ev.Evaluate(f.PageSize <= 100, "page_size", "must be a max of 100")
	if f.Sort != nil {
		ev.Evaluate(slices.Contains(*f.SafeSortList, *f.Sort), "sort", "invalid sort value")
	}
}

func (f *Filter) SortColumn() string {
	for _, safeValue := range *f.SafeSortList {
		if *f.Sort == safeValue {
			return strings.TrimPrefix(*f.Sort, "-")
		}
	}
	panic(fmt.Sprintf("unsafe sort param %v", f.Sort))
}

func (f *Filter) SortDirection() string {
	if strings.HasPrefix(*f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func (f *Filter) Limit() int {
	return f.PageSize
}

func (f *Filter) Offset() int {
	return (f.Page - 1) * f.PageSize
}

type Metadata struct {
	CurrentPage  int `json:"currentPage,omitempty"`
	PageSize     int `json:"pageSize,omitempty"`
	FirstPage    int `json:"firstPage,omitempty"`
	LastPage     int `json:"lastPage,omitempty"`
	TotalRecords int `json:"totalRecords,omitempty"`
}

func CalculateMetadata(totalRecords, pageSize, currentPage int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}
	return Metadata{
		CurrentPage:  currentPage,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
