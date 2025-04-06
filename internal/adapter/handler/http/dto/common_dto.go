// internal/adapter/handler/http/dto/common_dto.go
package dto

// PaginatedResponseDTO defines the standard structure for paginated API responses.
type PaginatedResponseDTO struct {
	Data       interface{} `json:"data"`       // The slice of items for the current page (e.g., []AudioTrackResponseDTO)
	Total      int         `json:"total"`      // Total number of items matching the query
	Limit      int         `json:"limit"`      // The limit used for this page
	Offset     int         `json:"offset"`     // The offset used for this page
	Page       int         `json:"page"`       // Current page number (1-based)
	TotalPages int         `json:"totalPages"` // Total number of pages
} 