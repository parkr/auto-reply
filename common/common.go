// common is a library I made because I was lazy and didn't know better.
package common

import (
	"fmt"
	"net/http"

	"github.com/google/go-github/github"
)

func SliceLookup(data []string) map[string]bool {
	mapping := map[string]bool{}
	for _, datum := range data {
		mapping[datum] = true
	}
	return mapping
}

func ErrorFromResponse(res *github.Response, err error) error {
	if err != nil {
		return err
	}

	if res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unexpected error code: %d", res.StatusCode)
	}

	return nil
}
