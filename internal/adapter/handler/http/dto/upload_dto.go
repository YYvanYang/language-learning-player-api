// internal/adapter/handler/http/dto/upload_dto.go
package dto

// === Single Upload DTOs ===

// RequestUploadRequestDTO defines the JSON body for requesting an upload URL.
type RequestUploadRequestDTO struct {
	Filename    string `json:"filename" validate:"required"`
	ContentType string `json:"contentType" validate:"required"` // e.g., "audio/mpeg"
}

// RequestUploadResponseDTO defines the JSON response after requesting an upload URL.
type RequestUploadResponseDTO struct {
	UploadURL string `json:"uploadUrl"` // The presigned PUT URL
	ObjectKey string `json:"objectKey"` // The key the client should use/report back
}

// CompleteUploadInputDTO defines the JSON body for finalizing an upload
// and creating the audio track metadata record.
type CompleteUploadInputDTO struct {
	ObjectKey     string   `json:"objectKey" validate:"required"`
	Title         string   `json:"title" validate:"required,max=255"`
	Description   string   `json:"description"`
	LanguageCode  string   `json:"languageCode" validate:"required"`
	Level         string   `json:"level" validate:"omitempty,oneof=A1 A2 B1 B2 C1 C2 NATIVE"` // Allow empty or valid level
	DurationMs    int64    `json:"durationMs" validate:"required,gt=0"`                       // Duration in Milliseconds, must be positive
	IsPublic      bool     `json:"isPublic"`                                                  // Defaults to false if omitted? Define behavior.
	Tags          []string `json:"tags"`
	CoverImageURL *string  `json:"coverImageUrl" validate:"omitempty,url"`
}

// === Batch Upload DTOs ===

// BatchRequestUploadInputItemDTO represents a single file in the batch request for URLs.
type BatchRequestUploadInputItemDTO struct {
	Filename    string `json:"filename" validate:"required"`
	ContentType string `json:"contentType" validate:"required"` // e.g., "audio/mpeg"
}

// BatchRequestUploadInputRequestDTO is the request body for requesting multiple upload URLs.
type BatchRequestUploadInputRequestDTO struct {
	Files []BatchRequestUploadInputItemDTO `json:"files" validate:"required,min=1,dive"` // Ensure at least one file, validate each item
}

// BatchRequestUploadInputResponseItemDTO represents the response for a single file URL request.
type BatchRequestUploadInputResponseItemDTO struct {
	OriginalFilename string `json:"originalFilename"` // Helps client match response to request
	ObjectKey        string `json:"objectKey"`        // The generated object key for this file
	UploadURL        string `json:"uploadUrl"`        // The presigned PUT URL for this file
	Error            string `json:"error,omitempty"`  // Error message if URL generation failed for this item
}

// BatchRequestUploadInputResponseDTO is the response body containing results for multiple URL requests.
type BatchRequestUploadInputResponseDTO struct {
	Results []BatchRequestUploadInputResponseItemDTO `json:"results"`
}

// BatchCompleteUploadItemDTO represents metadata for one successfully uploaded file in the batch completion request.
type BatchCompleteUploadItemDTO struct {
	ObjectKey     string   `json:"objectKey" validate:"required"`
	Title         string   `json:"title" validate:"required,max=255"`
	Description   string   `json:"description"`
	LanguageCode  string   `json:"languageCode" validate:"required"`
	Level         string   `json:"level" validate:"omitempty,oneof=A1 A2 B1 B2 C1 C2 NATIVE"`
	DurationMs    int64    `json:"durationMs" validate:"required,gt=0"`
	IsPublic      bool     `json:"isPublic"`
	Tags          []string `json:"tags"`
	CoverImageURL *string  `json:"coverImageUrl" validate:"omitempty,url"`
}

// BatchCompleteUploadInputDTO is the request body for finalizing multiple uploads.
type BatchCompleteUploadInputDTO struct {
	Tracks []BatchCompleteUploadItemDTO `json:"tracks" validate:"required,min=1,dive"` // Ensure at least one track, validate each item
}

// BatchCompleteUploadResponseItemDTO represents the processing result for a single item in the batch completion.
type BatchCompleteUploadResponseItemDTO struct {
	ObjectKey string `json:"objectKey"`         // Identifies the item
	Success   bool   `json:"success"`           // Whether processing this item succeeded
	TrackID   string `json:"trackId,omitempty"` // The ID of the created track if successful
	Error     string `json:"error,omitempty"`   // Error message if processing failed for this item
}

// BatchCompleteUploadResponseDTO is the overall response for the batch completion request.
// Note: This will likely be a 2xx status even if some items failed validation before DB commit,
// or a 4xx/5xx if the overall transaction failed. The details are in the items.
// Consider using 207 Multi-Status if you want to represent partial success explicitly at the HTTP level,
// but for atomicity, usually it's all-or-nothing for the DB part.
type BatchCompleteUploadResponseDTO struct {
	Results []BatchCompleteUploadResponseItemDTO `json:"results"`
}
