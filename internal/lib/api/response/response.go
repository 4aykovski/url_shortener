package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusOK                     = "OK"
	StatusError                  = "Error"
	InternalErrorMessage         = "internal error"
	DecodeErrorMessage           = "failed to decode request body"
	InvalidRequestErrorMessage   = "invalid request"
	WrongCredentialsErrorMessage = "wrong credentials"
	UnauthorizedErrorMessage     = "unauthorized"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func InternalError() Response {
	return Error(InternalErrorMessage)
}

func DecodeError() Response {
	return Error(DecodeErrorMessage)
}

func InvalidRequestError() Response {
	return Error(InvalidRequestErrorMessage)
}

func WrongCredentialsError() Response {
	return Error(WrongCredentialsErrorMessage)
}

func UnauthorizedError() Response {
	return Error(UnauthorizedErrorMessage)
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid URL", err.Field()))
		case "min":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s must be longer than %s symbols", err.Field(), err.Param()))
		case "max":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s must be smaller than %s symbols", err.Field(), err.Param()))
		case "containsany":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s must contains any of special character", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}
