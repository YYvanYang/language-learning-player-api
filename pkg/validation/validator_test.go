// ============================================
// FILE: pkg/validation/validator_test.go
// ============================================
package validation_test

import (
	"testing"

	"github.com/yvanyang/language-learning-player-backend/pkg/validation" // Adjust
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	RequiredField string `json:"reqField" validate:"required"`
	EmailField    string `json:"email_field" validate:"required,email"`
	NumberField   int    `json:"numField" validate:"gte=1,lte=10"`
	OptionalField string `json:"optField,omitempty" validate:"omitempty,url"`
}

func TestValidator_ValidateStruct_Success(t *testing.T) {
	validator := validation.New()
	validStruct := TestStruct{
		RequiredField: "value",
		EmailField:    "test@example.com",
		NumberField:   5,
		OptionalField: "http://example.com",
	}

	err := validator.ValidateStruct(validStruct)
	assert.NoError(t, err)

	// Test with optional field empty
	validStructOptionalEmpty := TestStruct{
		RequiredField: "value",
		EmailField:    "test@example.com",
		NumberField:   5,
		OptionalField: "", // Empty optional field is valid
	}
	err = validator.ValidateStruct(validStructOptionalEmpty)
	assert.NoError(t, err)
}

func TestValidator_ValidateStruct_Failure(t *testing.T) {
	validator := validation.New()

	testCases := []struct {
		name        string
		input       TestStruct
		errContains []string // Substrings expected in the error message
	}{
		{
			name: "Required Field Missing",
			input: TestStruct{
				EmailField:  "test@example.com",
				NumberField: 5,
			},
			errContains: []string{"'reqField' failed validation on the 'required' rule"},
		},
		{
			name: "Invalid Email",
			input: TestStruct{
				RequiredField: "value",
				EmailField:    "not-an-email",
				NumberField:   5,
			},
			errContains: []string{"'email_field' failed validation on the 'email' rule"},
		},
		{
			name: "Number Too Low",
			input: TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				NumberField:   0,
			},
			errContains: []string{"'numField' failed validation on the 'gte' rule"},
		},
		{
			name: "Number Too High",
			input: TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				NumberField:   11,
			},
			errContains: []string{"'numField' failed validation on the 'lte' rule"},
		},
		{
			name: "Invalid URL for Optional Field",
			input: TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				NumberField:   5,
				OptionalField: "not a url", // Invalid URL when present
			},
			errContains: []string{"'optField' failed validation on the 'url' rule"},
		},
		{
			name: "Multiple Errors",
			input: TestStruct{
				EmailField:  "invalid",
				NumberField: 15,
			},
			errContains: []string{
				"'reqField' failed validation on the 'required' rule",
				"'email_field' failed validation on the 'email' rule",
				"'numField' failed validation on the 'lte' rule",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateStruct(tc.input)
			require.Error(t, err, "Expected a validation error")
			errString := err.Error()
            assert.Contains(t, errString, "validation failed: ")
			for _, substr := range tc.errContains {
				assert.Contains(t, errString, substr, "Error message should contain specific rule failure")
			}
		})
	}
}