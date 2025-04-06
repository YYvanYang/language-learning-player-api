// ============================================
// FILE: pkg/pagination/pagination_test.go
// ============================================
package pagination_test

import (
	"testing"

	"github.com/yvanyang/language-learning-player-backend/pkg/pagination"
	"github.com/stretchr/testify/assert"
)

func TestNewPage(t *testing.T) {
	testCases := []struct {
		name           string
		pageNum        int
		pageSize       int
		expectedLimit  int
		expectedOffset int
	}{
		{"Page 1, Default Size", 1, 0, pagination.DefaultLimit, 0},
		{"Page 1, Specific Size", 1, 15, 15, 0},
		{"Page 2, Default Size", 2, 0, pagination.DefaultLimit, pagination.DefaultLimit},
		{"Page 3, Size 15", 3, 15, 15, 30},
		{"Zero Page Size", 1, 0, pagination.DefaultLimit, 0},
		{"Negative Page Size", 2, -5, pagination.DefaultLimit, pagination.DefaultLimit},
		{"Exceed Max Page Size", 1, pagination.MaxLimit + 50, pagination.MaxLimit, 0},
		{"Zero Page Num", 0, 25, 25, 0},          // Page 0 treated as Page 1
		{"Negative Page Num", -1, 25, 25, 0},     // Negative page treated as Page 1
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			page := pagination.NewPage(tc.pageNum, tc.pageSize)
			assert.Equal(t, tc.expectedLimit, page.Limit)
			assert.Equal(t, tc.expectedOffset, page.Offset)
		})
	}
}

func TestNewPageFromOffset(t *testing.T) {
	testCases := []struct {
		name           string
		limit          int
		offset         int
		expectedLimit  int
		expectedOffset int
	}{
		{"Valid", 10, 20, 10, 20},
		{"Default Limit", 0, 10, pagination.DefaultLimit, 10},
		{"Max Limit", pagination.MaxLimit + 1, 0, pagination.MaxLimit, 0},
		{"Negative Limit", -5, 5, pagination.DefaultLimit, 5},
		{"Zero Offset", 15, 0, 15, 0},
		{"Negative Offset", 15, -10, 15, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			page := pagination.NewPageFromOffset(tc.limit, tc.offset)
			assert.Equal(t, tc.expectedLimit, page.Limit)
			assert.Equal(t, tc.expectedOffset, page.Offset)
		})
	}
}


func TestNewPaginatedResponse(t *testing.T) {
	data := []string{"a", "b"}
	total := 25
	pageParams := pagination.Page{Limit: 10, Offset: 10} // Requesting page 2 (10 items/page)

	resp := pagination.NewPaginatedResponse(data, total, pageParams)

	assert.Equal(t, data, resp.Data)
	assert.Equal(t, total, resp.Total)
	assert.Equal(t, 10, resp.Limit)
	assert.Equal(t, 10, resp.Offset)
	assert.Equal(t, 2, resp.Page) // Offset 10 with limit 10 is page 2
	assert.Equal(t, 3, resp.TotalPages) // Ceil(25 / 10) = 3
}

func TestNewPaginatedResponse_EdgeCases(t *testing.T) {
	// Zero total
	resp := pagination.NewPaginatedResponse([]int{}, 0, pagination.Page{Limit: 10, Offset: 0})
	assert.Equal(t, 0, resp.Total)
	assert.Equal(t, 10, resp.Limit) // Limit stays as requested/defaulted
	assert.Equal(t, 0, resp.Offset)
	assert.Equal(t, 1, resp.Page) // Page is 1 even with 0 items
	assert.Equal(t, 0, resp.TotalPages)

	// Offset exceeding total
	resp = pagination.NewPaginatedResponse([]string{}, 15, pagination.Page{Limit: 10, Offset: 20})
	assert.Equal(t, 15, resp.Total)
	assert.Equal(t, 10, resp.Limit)
	assert.Equal(t, 20, resp.Offset)
	assert.Equal(t, 3, resp.Page) // Offset 20 with limit 10 is page 3
	assert.Equal(t, 2, resp.TotalPages) // Ceil(15 / 10) = 2
}

func TestNewPaginatedResponse_ConstraintApplication(t *testing.T) {
    // Test if constraints are re-applied
    data := []string{"a"}
    total := 5
    // Provide invalid page params
    pageParams := pagination.Page{Limit: pagination.MaxLimit + 10, Offset: -5}

    resp := pagination.NewPaginatedResponse(data, total, pageParams)

    assert.Equal(t, pagination.MaxLimit, resp.Limit, "Limit should be constrained")
    assert.Equal(t, 0, resp.Offset, "Offset should be non-negative")
    assert.Equal(t, 1, resp.Page, "Page should be 1 for offset 0")
    assert.Equal(t, 1, resp.TotalPages, "Total pages should be Ceil(5 / MaxLimit)")
}