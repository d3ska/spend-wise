package model

// AppError is a domain error that carries a machine-readable code alongside
// a human-readable message. It implements the error interface and supports
// identity comparison via errors.Is().
type AppError struct {
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

// Is supports errors.Is() by comparing the Code field.
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// NewAppError creates an AppError with the given code and message.
func NewAppError(code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}
