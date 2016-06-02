package common

import (
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-github/github"
)

func TestErrorFromResponse(t *testing.T) {
	transportErr := errors.New("something terrible happened!")
	resError := errors.New("unexpected error code: 404")
	httpRes := &http.Response{StatusCode: http.StatusNotFound}
	res := &github.Response{httpRes, 0, 0, 0, 0, github.Rate{}}

	err := ErrorFromResponse(res, transportErr)
	if err != transportErr {
		t.Fatalf("expected %v, got: %v", transportErr, err)
	}

	err = ErrorFromResponse(res, nil)
	if err.Error() != resError.Error() {
		t.Fatalf("expected %v, got: %v", resError, err)
	}
}
