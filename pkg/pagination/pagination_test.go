// ============================================
// FILE: pkg/pagination/pagination_test.go
// ============================================

package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPage(t *testing.T) {
	tests := []struct {
		name         string
		pageNum      int
		pageSize     int
		expectedPage Page
	}{
		{"Valid page 1", 1, 15, Page{Limit: 15, Offset: 0}},
		{"Valid page 3", 3, 10, Page{Limit: 10, Offset: 20}},
		{"Zero page size", 2, 0, Page{Limit: DefaultLimit, Offset: DefaultLimit}},      // Uses default limit for offset calculation
		{"Negative page size", 2, -5, Page{Limit: DefaultLimit, Offset: DefaultLimit}}, // Uses default limit
		{"Page size exceeds max", 1, MaxLimit + 10, Page{Limit: MaxLimit, Offset: 0}},
		{"Page number 0", 0, 20, Page{Limit: 20, Offset: 0}},         // Treated as page 1
		{"Page number negative", -1, 20, Page{Limit: 20, Offset: 0}}, // Treated as page 1
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPage(tt.pageNum, tt.pageSize)
			assert.Equal(t, tt.expectedPage, p)
		})
	}
}

func TestNewPageFromOffset(t *testing.T) {
	tests := []struct {
		name         string
		limit        int
		offset       int
		expectedPage Page
	}{
		{"Valid limit and offset", 15, 30, Page{Limit: 15, Offset: 30}},
		{"Zero limit", 0, 50, Page{Limit: DefaultLimit, Offset: 50}},
		{"Negative limit", -10, 50, Page{Limit: DefaultLimit, Offset: 50}},
		{"Limit exceeds max", MaxLimit + 5, 10, Page{Limit: MaxLimit, Offset: 10}},
		{"Zero offset", 25, 0, Page{Limit: 25, Offset: 0}},
		{"Negative offset", 25, -10, Page{Limit: 25, Offset: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPageFromOffset(tt.limit, tt.offset)
			assert.Equal(t, tt.expectedPage, p)
		})
	}
}

func TestNewPaginatedResponse(t *testing.T) {
	type sampleData struct {
		ID int
	}
	dataPage1 := []sampleData{{ID: 1}, {ID: 2}}
	dataPageLast := []sampleData{{ID: 5}}

	tests := []struct {
		name        string
		data        interface{}
		total       int
		pageParams  Page
		expectedRes PaginatedResponse
	}{
		{
			name:       "First page",
			data:       dataPage1,
			total:      5,
			pageParams: Page{Limit: 2, Offset: 0},
			expectedRes: PaginatedResponse{
				Data:       dataPage1,
				Total:      5,
				Limit:      2,
				Offset:     0,
				Page:       1, // (0 / 2) + 1
				TotalPages: 3, // ceil(5 / 2)
			},
		},
		{
			name:       "Middle page",
			data:       dataPage1, // Assuming data for page 2 is []sampleData{{ID: 3}, {ID: 4}}
			total:      5,
			pageParams: Page{Limit: 2, Offset: 2},
			expectedRes: PaginatedResponse{
				Data:       dataPage1, // Pass the actual data for the page being tested
				Total:      5,
				Limit:      2,
				Offset:     2,
				Page:       2, // (2 / 2) + 1
				TotalPages: 3,
			},
		},
		{
			name:       "Last page",
			data:       dataPageLast,
			total:      5,
			pageParams: Page{Limit: 2, Offset: 4},
			expectedRes: PaginatedResponse{
				Data:       dataPageLast,
				Total:      5,
				Limit:      2,
				Offset:     4,
				Page:       3, // (4 / 2) + 1
				TotalPages: 3,
			},
		},
		{
			name:       "Total items less than limit",
			data:       dataPage1,
			total:      2,
			pageParams: Page{Limit: 10, Offset: 0},
			expectedRes: PaginatedResponse{
				Data:       dataPage1,
				Total:      2,
				Limit:      10,
				Offset:     0,
				Page:       1,
				TotalPages: 1, // ceil(2 / 10)
			},
		},
		{
			name:       "Zero total items",
			data:       []sampleData{},
			total:      0,
			pageParams: Page{Limit: 10, Offset: 0},
			expectedRes: PaginatedResponse{
				Data:       []sampleData{},
				Total:      0,
				Limit:      10,
				Offset:     0,
				Page:       1,
				TotalPages: 0,
			},
		},
		{
			name:       "Invalid page params (default limit)",
			data:       dataPage1,
			total:      5,
			pageParams: Page{Limit: 0, Offset: 0}, // Should use DefaultLimit
			expectedRes: PaginatedResponse{
				Data:       dataPage1,
				Total:      5,
				Limit:      DefaultLimit,
				Offset:     0,
				Page:       1,
				TotalPages: 1, // ceil(5 / DefaultLimit)
			},
		},
		{
			name:       "Invalid page params (max limit)",
			data:       dataPage1,
			total:      200,
			pageParams: Page{Limit: MaxLimit + 10, Offset: 0}, // Should use MaxLimit
			expectedRes: PaginatedResponse{
				Data:       dataPage1,
				Total:      200,
				Limit:      MaxLimit,
				Offset:     0,
				Page:       1,
				TotalPages: 2, // ceil(200 / MaxLimit)
			},
		},
		{
			name:       "Invalid page params (negative offset)",
			data:       dataPage1,
			total:      5,
			pageParams: Page{Limit: 10, Offset: -10}, // Should use Offset 0
			expectedRes: PaginatedResponse{
				Data:       dataPage1,
				Total:      5,
				Limit:      10,
				Offset:     0,
				Page:       1, // (0 / 10) + 1
				TotalPages: 1, // ceil(5 / 10)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewPaginatedResponse(tt.data, tt.total, tt.pageParams)
			assert.Equal(t, tt.expectedRes, res)
		})
	}
}
