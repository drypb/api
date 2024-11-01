// Package validator helps the API validate client input.
package validator

import "regexp"

var (
	// EmailRX is a regular expression to identify valid emails.
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Validator struct {
	Errors map[string]string // Errors store all the error messages [Validator] catches.
}

// New creates a new Validator
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid return true if the [Validator] catched no errors.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// Check adds a new error to [Errors] if not ok.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// AddError adds a new error to |Validator.Errors|.
func (v *Validator) AddError(key, message string) {
	_, exists := v.Errors[key]
	if !exists {
		v.Errors[key] = message
	}
}

// PermittedValue returns true if the value is in permittedValues.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	for _, permittedValue := range permittedValues {
		if value == permittedValue {
			return true
		}
	}
	return false
}

// Matches returns true if the regular expression matches.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique returns true if there are no repetitions in values.
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}
