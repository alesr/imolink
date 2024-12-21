package errors

import "fmt"

type ErrorType string

const (
	ErrorTypeAPI           ErrorType = "api_error"
	ErrorTypeTimeout       ErrorType = "timeout_error"
	ErrorTypeValidation    ErrorType = "validation_error"
	ErrorTypeTranscription ErrorType = "transcription_error"
)

type Error struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func New(errType ErrorType, message string, err error) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}
