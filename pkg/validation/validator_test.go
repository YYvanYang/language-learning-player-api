package validation

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	RequiredField string    `json:"reqField" validate:"required"`
	EmailField    string    `json:"emailField" validate:"required,email"`
	UUIDField     string    `json:"uuidField" validate:"required,uuid"`
	MinLenField   string    `json:"minLenField" validate:"required,min=5"`
	MaxLenField   string    `json:"maxLenField" validate:"required,max=10"`
	OneOfField    string    `json:"oneOfField" validate:"required,oneof=A B C"`
	OptionalField *string   `json:"optField" validate:"omitempty,min=3"` // Optional, but min length if present
	DiveField     []SubTest `json:"diveField" validate:"required,dive"`
}

type SubTest struct {
	SubField string `json:"subField" validate:"required,alpha"`
}

func TestValidator_ValidateStruct(t *testing.T) {
	v := New()
	validUUID := uuid.NewString()

	tests := []struct {
		name      string
		input     interface{}
		expectErr bool
		errSubstr []string // Substrings expected in the error message
	}{
		{
			name: "Valid struct",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     validUUID,
				MinLenField:   "12345",
				MaxLenField:   "1234567890",
				OneOfField:    "A",
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: false,
		},
		{
			name: "Missing required field",
			input: &TestStruct{
				// RequiredField missing
				EmailField:  "test@example.com",
				UUIDField:   validUUID,
				MinLenField: "12345",
				MaxLenField: "1234567890",
				OneOfField:  "B",
				DiveField:   []SubTest{{SubField: "Alpha"}},
			},
			expectErr: true,
			errSubstr: []string{"'reqField' failed validation on the 'required' rule"},
		},
		{
			name: "Invalid email format",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "not-an-email", // Invalid
				UUIDField:     validUUID,
				MinLenField:   "12345",
				MaxLenField:   "1234567890",
				OneOfField:    "C",
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: true,
			errSubstr: []string{"'emailField' failed validation on the 'email' rule"},
		},
		{
			name: "Invalid UUID format",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     "not-a-uuid", // Invalid
				MinLenField:   "12345",
				MaxLenField:   "1234567890",
				OneOfField:    "A",
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: true,
			errSubstr: []string{"'uuidField' failed validation on the 'uuid' rule"},
		},
		{
			name: "Min length failure",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     validUUID,
				MinLenField:   "123", // Too short
				MaxLenField:   "1234567890",
				OneOfField:    "B",
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: true,
			errSubstr: []string{"'minLenField' failed validation on the 'min' rule"},
		},
		{
			name: "Max length failure",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     validUUID,
				MinLenField:   "12345",
				MaxLenField:   "12345678901", // Too long
				OneOfField:    "C",
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: true,
			errSubstr: []string{"'maxLenField' failed validation on the 'max' rule"},
		},
		{
			name: "OneOf failure",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     validUUID,
				MinLenField:   "12345",
				MaxLenField:   "1234567890",
				OneOfField:    "D", // Not in A, B, C
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: true,
			errSubstr: []string{"'oneOfField' failed validation on the 'oneof' rule"},
		},
		{
			name: "Valid optional field (present)",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     validUUID,
				MinLenField:   "12345",
				MaxLenField:   "1234567890",
				OneOfField:    "A",
				OptionalField: func() *string { s := "abc"; return &s }(), // Valid length
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: false,
		},
		{
			name: "Valid optional field (absent)",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     validUUID,
				MinLenField:   "12345",
				MaxLenField:   "1234567890",
				OneOfField:    "B",
				OptionalField: nil, // Absent, should pass omitempty
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: false,
		},
		{
			name: "Invalid optional field (present but too short)",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     validUUID,
				MinLenField:   "12345",
				MaxLenField:   "1234567890",
				OneOfField:    "C",
				OptionalField: func() *string { s := "ab"; return &s }(), // Too short
				DiveField:     []SubTest{{SubField: "Alpha"}},
			},
			expectErr: true,
			errSubstr: []string{"'optField' failed validation on the 'min' rule"},
		},
		{
			name: "Dive validation failure (sub-struct)",
			input: &TestStruct{
				RequiredField: "value",
				EmailField:    "test@example.com",
				UUIDField:     validUUID,
				MinLenField:   "12345",
				MaxLenField:   "1234567890",
				OneOfField:    "A",
				DiveField:     []SubTest{{SubField: "Not1Alpha"}}, // Invalid sub-field
			},
			expectErr: true,
			errSubstr: []string{"'subField' failed validation on the 'alpha' rule"}, // Note: Field name comes from sub-struct
		},
		{
			name: "Multiple errors",
			input: &TestStruct{
				// RequiredField missing
				EmailField:  "not-an-email",
				UUIDField:   "not-uuid",
				MinLenField: "123",
				MaxLenField: "12345678901",
				OneOfField:  "D",
				DiveField:   []SubTest{{SubField: "Not1Alpha"}},
			},
			expectErr: true,
			errSubstr: []string{
				"'reqField' failed validation on the 'required' rule",
				"'emailField' failed validation on the 'email' rule",
				"'uuidField' failed validation on the 'uuid' rule",
				"'minLenField' failed validation on the 'min' rule",
				"'maxLenField' failed validation on the 'max' rule",
				"'oneOfField' failed validation on the 'oneof' rule",
				"'subField' failed validation on the 'alpha' rule",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateStruct(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
				if err != nil {
					for _, sub := range tt.errSubstr {
						assert.Contains(t, err.Error(), sub)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
