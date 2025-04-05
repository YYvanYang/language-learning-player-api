// internal/adapter/handler/http/dto/upload_dto.go
package dto

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

// CompleteUploadRequestDTO defines the JSON body for finalizing an upload
// and creating the audio track metadata record.
type CompleteUploadRequestDTO struct {
	ObjectKey     string   `json:"objectKey" validate:"required"`
	Title         string   `json:"title" validate:"required,max=255"`
	Description   string   `json:"description"`
	LanguageCode  string   `json:"languageCode" validate:"required"`
	Level         string   `json:"level" validate:"omitempty,oneof=A1 A2 B1 B2 C1 C2 NATIVE"` // Allow empty or valid level
	DurationMs    int64    `json:"durationMs" validate:"required,gt=0"` // Duration in Milliseconds, must be positive
	IsPublic      bool     `json:"isPublic"` // Defaults to false if omitted? Define behavior.
	Tags          []string `json:"tags"`
	CoverImageURL *string  `json:"coverImageUrl" validate:"omitempty,url"`
} 