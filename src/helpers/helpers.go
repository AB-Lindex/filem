package helpers

import (
	"fmt"
)

// StringError is a helper for converting strings to an `error`
type StringError string

// Error return the error-text for a StringError
func (err StringError) Error() string {
	return string(err)
}

// RequiredError is a helper for converting a fieldname to a 'is required'-error
type RequiredError string

// Error return the error-text for a RequiredError
func (name RequiredError) Error() string {
	return fmt.Sprintf("'%s' is required", string(name))
}

// InvalidError is a helper for converting a fieldname to a 'is invalid'-error
type InvalidError string

// Error return the error-text for a InvalidError
func (name InvalidError) Error() string {
	return fmt.Sprintf("'%s' is invalid", string(name))
}
