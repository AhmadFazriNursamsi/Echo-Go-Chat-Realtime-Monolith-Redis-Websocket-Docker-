package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// FormatValidationError bikin pesan error lebih singkat & rapi
func FormatValidationError(err error) []string {
	var errors []string
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			// Ambil nama field (contoh: RegisterUserDto.Email → jadi email)
			field := strings.ToLower(e.Field())

			switch e.Tag() {
			case "required":
				errors = append(errors, fmt.Sprintf("%s is required", field))
			case "email":
				errors = append(errors, "invalid email format")
			case "min":
				errors = append(errors, fmt.Sprintf("%s must be at least %s characters", field, e.Param()))
			default:
				errors = append(errors, fmt.Sprintf("%s is not valid", field))
			}
		}
	} else {
		errors = append(errors, err.Error())
	}
	return errors
}
