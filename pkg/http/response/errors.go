package utls

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Error struct {
	Errors map[string]interface{} `json:"errors"`
}

type ErrorResponse struct {
	Success    bool      `json:"success"`
	Message    string    `json:"message"`
	StatusCode int       `json:"status_code"`
	Data       ErrorData `json:"data"`
}
type ErrorData struct {
	Error string `json:"error"`
}

func NewResponseError(message string, statusCode int, err error) ErrorResponse {
	return ErrorResponse{
		Success:    false,
		Message:    message,
		StatusCode: statusCode,
		Data: ErrorData{
			Error: err.Error(),
		},
	}
}

// add switch other variant
func NewError(err error) Error {
	e := Error{}
	e.Errors = make(map[string]interface{})
	switch v := err.(type) {
	default:
		e.Errors["body"] = v.Error()
	}
	return e
}

func NewValidatorError(err error) Error {
	e := Error{}
	e.Errors = make(map[string]interface{})
	errs := err.(validator.ValidationErrors)
	for _, v := range errs {
		e.Errors[v.Field()] = fmt.Sprintf("%v", v.Tag())
	}
	return e
}

func AccessForbidden() Error {
	e := Error{}
	e.Errors = make(map[string]interface{})
	e.Errors["body"] = "access forbidden"
	return e
}

func NotFound() Error {
	e := Error{}
	e.Errors = make(map[string]interface{})
	e.Errors["body"] = "resource not found"
	return e
}
