package serviceerror_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/protogalaxy/common/serviceerror"
)

func TestErrorResponseDecode(t *testing.T) {
	input := strings.NewReader(`{"status":400, "message":"msg", "error":"errcode"}`)
	err := serviceerror.Decode(input)
	if err == nil {
		t.Fatalf("Error value should always be returned")
	}
	serr, ok := err.(serviceerror.ErrorResponse)
	if !ok {
		t.Fatalf("The returned error is not ErrorResponse but got: %s", err)
	}
	if serr.StatusCode != http.StatusBadRequest {
		t.Fatalf("Unexpected response code: '%d'", serr.StatusCode)
	}
	if serr.ErrorCode != "errcode" {
		t.Fatalf("Unexpected error code: '%s'", serr.ErrorCode)
	}
	if serr.Message != "msg" {
		t.Fatalf("Unexpected response message: '%s'", serr.Message)
	}
	if serr.Cause != nil {
		t.Fatalf("There should be no cause in the decoded message but got: %s", serr.Cause)
	}
}

func TestErrorResponseEncode(t *testing.T) {
	cause := errors.New("err")
	errorResponse := serviceerror.InternalServerError("code", "msg", cause)
	res, err := json.Marshal(&errorResponse)
	if err != nil {
		t.Fatalf("Unexpected marshal error: %s", err)
	}
	result := serviceerror.Decode(bytes.NewReader(res))
	er, ok := result.(serviceerror.ErrorResponse)
	if !ok {
		t.Fatalf("Unexpected error occured: %s", result)
	}
	if er.StatusCode != http.StatusInternalServerError {
		t.Errorf("Unexpected status code: %d", er.StatusCode)
	}
	if er.ErrorCode != "code" {
		t.Errorf("Unexpected error code: %s", er.ErrorCode)
	}
	if er.Message != "msg" {
		t.Errorf("Unexpected message: %s", er.Message)
	}
	if er.Cause != nil {
		t.Errorf("Unexpected cause: %v", er.Cause)
	}
}

func TestErrorResponseRequiredStatusCode(t *testing.T) {
	input := strings.NewReader(`{message":"msg", "error":"errcode"}`)
	err := serviceerror.Decode(input)
	if _, ok := err.(serviceerror.ErrorResponse); ok {
		t.Fatalf("Error field status is required")
	}
}

func TestErrorResponseRequiredErrorCode(t *testing.T) {
	input := strings.NewReader(`{message":"msg", "code":"100"}`)
	err := serviceerror.Decode(input)
	if _, ok := err.(serviceerror.ErrorResponse); ok {
		t.Fatalf("Error field error is required")
	}
}
