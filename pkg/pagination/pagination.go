// pkg/pagination/pagination.go
package pagination

// DefaultLimit defines the default number of items per page.
const DefaultLimit = 20
// MaxLimit defines the maximum number of items allowed per page.
const MaxLimit = 100

// Page represents pagination parameters.
type Page struct {
	Limit  int `json:"limit" schema:"limit"`   // Number of items per page
	Offset int `json:"offset" schema:"offset"` // Number of items to skip
}

// GetLimit returns the pagination limit, applying defaults and maximums.
func (p *Page) GetLimit() int {
	if p.Limit <= 0 {
		return DefaultLimit
	}
	if p.Limit > MaxLimit {
		return MaxLimit
	}
	return p.Limit
}

// GetOffset returns the pagination offset, ensuring it's non-negative.
func (p *Page) GetOffset() int {
	if p.Offset < 0 {
		return 0
	}
	return p.Offset
} 