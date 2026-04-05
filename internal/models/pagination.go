package models

import "fmt"

type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
	TotalCount int `json:"total_count"`
}

func (p Pagination) Summary() string {
	return fmt.Sprintf("Page %d of %d (%d total)", p.Page, p.TotalPages, p.TotalCount)
}

func (p Pagination) HasNext() bool {
	return p.Page < p.TotalPages
}

func (p Pagination) HasPrev() bool {
	return p.Page > 1
}
