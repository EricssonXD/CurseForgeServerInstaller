package errors

import "fmt"

// McServerError is the base error type for all mcserver errors.
type McServerError struct {
	Message string
}

func (e *McServerError) Error() string {
	return e.Message
}

// UserFacingError wraps errors that should be shown to the user as-is.
type UserFacingError struct {
	McServerError
}

func NewUserFacingError(msg string) *UserFacingError {
	return &UserFacingError{McServerError{Message: msg}}
}

func NewUserFacingErrorf(format string, args ...any) *UserFacingError {
	return &UserFacingError{McServerError{Message: fmt.Sprintf(format, args...)}}
}

// MissingApiKeyError indicates the CurseForge API key is not configured.
type MissingApiKeyError struct {
	McServerError
}

func NewMissingApiKeyError(msg string) *MissingApiKeyError {
	return &MissingApiKeyError{McServerError{Message: msg}}
}

// InvalidApiKeyError indicates the API key was rejected by CurseForge.
type InvalidApiKeyError struct {
	McServerError
}

func NewInvalidApiKeyError(msg string) *InvalidApiKeyError {
	return &InvalidApiKeyError{McServerError{Message: msg}}
}

// IsUserFacing checks if an error should be displayed directly to the user.
func IsUserFacing(err error) bool {
	switch err.(type) {
	case *UserFacingError, *MissingApiKeyError, *InvalidApiKeyError:
		return true
	default:
		return false
	}
}
