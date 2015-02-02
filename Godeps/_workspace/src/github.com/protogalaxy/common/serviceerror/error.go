package serviceerror

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type ErrorResponse struct {
	StatusCode int    `json:"status"`
	ErrorCode  string `json:"error"`
	Message    string `json:"message,omitempty"`
	Cause      error  `json:"-"`
}

func (e ErrorResponse) Error() string {
	err := e.ErrorCode
	if e.Message != "" {
		err += ": " + e.Message
	}
	if e.Cause != nil {
		err += ": " + e.Cause.Error()
	}
	return err
}

func InternalServerError(code, msg string, err error) ErrorResponse {
	return ErrorResponse{
		StatusCode: http.StatusInternalServerError,
		ErrorCode:  code,
		Message:    msg,
		Cause:      err,
	}
}

func BadRequest(code, msg string) ErrorResponse {
	return ErrorResponse{
		StatusCode: http.StatusBadRequest,
		ErrorCode:  code,
		Message:    msg,
	}
}

func Decode(body io.Reader) error {
	var response ErrorResponse
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&response); err != nil {
		return fmt.Errorf("Error decoding ErrorResponse: %s", err)
	}
	if response.StatusCode == 0 {
		return errors.New("Missing field status")
	}
	if response.ErrorCode == "" {
		return errors.New("Missing field error")
	}
	return response
}
