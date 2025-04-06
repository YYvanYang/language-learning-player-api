// pkg/pagination/pagination.go
package pagination

import (
	"math"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// Page represents pagination parameters from a request.
type Page struct {
	Limit  int // Number of items per page
	Offset int // Number of items to skip
}

// NewPage creates a Page struct from limit and offset inputs, applying defaults and constraints.
// pageNum is 1-based, pageSize is the number of items per page.
func NewPage(pageNum, pageSize int) Page {
	limit := pageSize
	if limit <= 0 {
		limit = DefaultLimit // Apply default limit if invalid or zero
	}
	if limit > MaxLimit {
		limit = MaxLimit // Apply max limit constraint
	}

	offset := 0
	if pageNum > 1 {
		offset = (pageNum - 1) * limit // Calculate offset based on 1-based page number
	}
	// Ensure offset is not negative (though calculation above shouldn't make it negative if pageNum >= 1)
	if offset < 0 {
		offset = 0
	}


	return Page{
		Limit:  limit,
		Offset: offset,
	}
}

// NewPageFromOffset creates a Page struct directly from limit and offset, applying defaults and constraints.
func NewPageFromOffset(limit, offset int) Page {
    pLimit := limit
	if pLimit <= 0 {
		pLimit = DefaultLimit
	}
	if pLimit > MaxLimit {
		pLimit = MaxLimit
	}

    pOffset := offset
    if pOffset < 0 {
        pOffset = 0
    }

	return Page{
		Limit:  pLimit,
		Offset: pOffset,
	}
}


// PaginatedResponse defines the standard structure for paginated API responses.
// The Data field should hold a slice of the actual response DTOs (e.g., []dto.AudioTrackResponseDTO).
type PaginatedResponse struct {
	Data       interface{} `json:"data"`        // The slice of items for the current page
	Total      int         `json:"total"`       // Total number of items matching the query
	Limit      int         `json:"limit"`       // The limit used for this page
	Offset     int         `json:"offset"`      // The offset used for this page
	Page       int         `json:"page"`        // Current page number (1-based)
	TotalPages int         `json:"totalPages"` // Total number of pages
}

// NewPaginatedResponse creates a PaginatedResponse struct.
// It calculates Page and TotalPages based on the provided data.
// - data: The slice of items for the current page (e.g., []*domain.Track).
// - total: The total number of items available across all pages.
// - pageParams: The Page struct (Limit, Offset) used to fetch this data.
func NewPaginatedResponse(data interface{}, total int, pageParams Page) PaginatedResponse {
	// Ensure limit and offset from pageParams are valid (apply constraints again just in case)
	limit := pageParams.Limit
	if limit <= 0 {
        limit = DefaultLimit
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}
    offset := pageParams.Offset
    if offset < 0 {
        offset = 0
    }

	totalPages := 0
	if total > 0 && limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

    currentPage := 1
    if limit > 0 {
        currentPage = (offset / limit) + 1
    }


	return PaginatedResponse{
		Data:       data,
		Total:      total,
		Limit:      limit, // Use the constrained limit
		Offset:     offset, // Use the constrained offset
		Page:       currentPage,
		TotalPages: totalPages,
	}
}