// pkg/validation/validator.go
package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the validator.v10 instance.
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator instance.
func New() *Validator {
	validate := validator.New()

	// Register a function to get the 'json' tag name for field names in errors.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// TODO: Register custom validation rules here if needed
	// Example: validate.RegisterValidation("customRule", customValidationFunc)

	return &Validator{validate: validate}
}

// ValidateStruct validates the given struct based on 'validate' tags.
// Returns a formatted error string if validation fails, otherwise nil.
func (v *Validator) ValidateStruct(s interface{}) error {
	err := v.validate.Struct(s)
	if err != nil {
		// Translate validation errors into a user-friendly message
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			// Use the 'json' tag name if available, otherwise use the field name
			fieldName := err.Field()
			// Construct a message based on the validation tag
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' failed validation on the '%s' rule", fieldName, err.Tag()))
			// Alternative: More user-friendly messages based on tag
			// msg := fmt.Sprintf("Invalid value for field %s (%s rule)", fieldName, err.Tag())
			// errorMessages = append(errorMessages, msg)
		}
		// Return a single error string joining all messages
		// Prepend with domain.ErrInvalidArgument? Maybe not, let handler decide mapping.
		return fmt.Errorf("validation failed: %s", strings.Join(errorMessages, "; "))
		// Alternative: return the original validator.ValidationErrors if handler needs more detail
		// return err
	}
	return nil
}

// TODO: Add ValidateVariable for single field validation if needed.
// func (v *Validator) ValidateVariable(i interface{}, tag string) error { ... }