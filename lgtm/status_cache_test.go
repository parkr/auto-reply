package lgtm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
	"github.com/stretchr/testify/assert"
)

func TestLgtmContext(t *testing.T) {
	cases := []struct {
		owner    string
		expected string
	}{
		{"deadbeef", "deadbeef/lgtm"},
		{"jekyll", "jekyll/lgtm"},
	}
	for _, test := range cases {
		assert.Equal(t, test.expected, lgtmContext(test.owner))
	}
}

func TestGetStatusInCache(t *testing.T) {
	setup() // server & client!
	defer teardown()
	context := &ctx.Context{GitHub: client}
	ref := newPRRef("o", "r", 1)
	expectedInfo := &statusInfo{
		lgtmers: []string{"@parkr"},
	}

	statusCache = statusMap{data: make(map[string]*statusInfo)}
	statusCache.data[ref.String()] = expectedInfo

	info, err := getStatus(context, ref)

	assert.NoError(t, err)
	assert.Equal(t, expectedInfo, info)
}

func TestGetStatusAPIPRError(t *testing.T) {
	setup() // server & client!
	defer teardown()
	statusCache = statusMap{data: make(map[string]*statusInfo)}
	context := &ctx.Context{GitHub: client}
	ref := newPRRef("o", "r", 1)
	handled := false

	mux.HandleFunc("/repos/o/r/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		handled = true
		testMethod(t, r, "GET")
		http.Error(w, "huh?", http.StatusNotFound)
	})

	info, err := getStatus(context, ref)

	assert.True(t, handled)
	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Nil(t, statusCache.data[ref.String()])
}

func TestGetStatusAPIStatusesError(t *testing.T) {
	setup() // server & client!
	defer teardown()
	statusCache = statusMap{data: make(map[string]*statusInfo)}
	context := &ctx.Context{GitHub: client}
	ref := newPRRef("o", "r", 1)
	expectedSHA := "abcdef"
	prHandled := false
	statusesHandled := false

	mux.HandleFunc("/repos/o/r/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		json.NewEncoder(w).Encode(&github.PullRequest{
			Number: github.Int(1),
			Head: &github.PullRequestBranch{
				Ref: github.String("blah:hi"),
				SHA: github.String(expectedSHA),
			},
		})
		prHandled = true
	})

	mux.HandleFunc("/repos/o/r/commits/"+expectedSHA+"/statuses", func(w http.ResponseWriter, r *http.Request) {
		statusesHandled = true
		http.Error(w, "huh?", http.StatusNotFound)
	})

	info, err := getStatus(context, ref)

	assert.True(t, prHandled, "the pull request should be fetched from the API")
	assert.True(t, statusesHandled, "the /statuses endpoint should be hit")
	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Nil(t, statusCache.data[ref.String()])
}

func TestGetStatusAPIStatusesNoneMatch(t *testing.T) {
	setup() // server & client!
	defer teardown()
	statusCache = statusMap{data: make(map[string]*statusInfo)}
	context := &ctx.Context{GitHub: client}
	ref := newPRRef("o", "r", 1)
	ref.Repo.Quorum = 1
	expectedSHA := "abcdef"
	prHandled := false
	statusesHandled := false

	mux.HandleFunc("/repos/o/r/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		json.NewEncoder(w).Encode(&github.PullRequest{
			Number: github.Int(1),
			Head: &github.PullRequestBranch{
				Ref: github.String("blah:hi"),
				SHA: github.String(expectedSHA),
			},
		})
		prHandled = true
	})

	mux.HandleFunc("/repos/o/r/commits/"+expectedSHA+"/statuses", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		json.NewEncoder(w).Encode([]github.RepoStatus{
			{Context: github.String("other/lgtm")},
		})
		statusesHandled = true
	})

	info, err := getStatus(context, ref)

	expectedStatus := &statusInfo{
		lgtmers: []string{},
		sha:     expectedSHA,
	}
	expectedStatus.repoStatus = expectedStatus.NewStatus(ref.Repo.Owner, ref.Repo.Quorum)

	assert.True(t, prHandled, "the pull request should be fetched from the API")
	assert.True(t, statusesHandled, "the /statuses endpoint should be hit")
	assert.NoError(t, err)
	assert.Equal(t, expectedStatus, info)
	assert.Equal(t, info, statusCache.data[ref.String()])
}

func TestGetStatusFromAPI(t *testing.T) {
	setup() // server & client!
	defer teardown()
	statusCache = statusMap{data: make(map[string]*statusInfo)}
	context := &ctx.Context{GitHub: client}
	ref := newPRRef("o", "r", 1)
	expectedSHA := "abcdef"
	expectedRepoStatus := &github.RepoStatus{
		Context:     github.String("o/lgtm"),
		Description: github.String("@parkr, @envygeeks, and @mattr- have approved this PR."),
	}
	prHandled := false
	statusesHandled := false

	mux.HandleFunc("/repos/o/r/pulls/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		json.NewEncoder(w).Encode(&github.PullRequest{
			Number: github.Int(1),
			Head: &github.PullRequestBranch{
				Ref: github.String("blah:hi"),
				SHA: github.String(expectedSHA),
			},
		})
		prHandled = true
	})

	mux.HandleFunc("/repos/o/r/commits/"+expectedSHA+"/statuses", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		json.NewEncoder(w).Encode([]github.RepoStatus{
			{Context: github.String("other/lgtm"), Description: github.String("no")},
			*expectedRepoStatus,
		})
		statusesHandled = true
	})

	info, err := getStatus(context, ref)

	expectedStatus := &statusInfo{
		lgtmers: []string{"@parkr", "@envygeeks", "@mattr-"},
		sha:     expectedSHA,
	}
	expectedStatus.repoStatus = expectedRepoStatus

	assert.True(t, prHandled, "the pull request should be fetched from the API")
	assert.True(t, statusesHandled, "the /statuses endpoint should be hit")
	assert.NoError(t, err)
	assert.Equal(t, expectedStatus, info)
	assert.Equal(t, expectedRepoStatus, info.repoStatus)
	assert.Equal(t, info, statusCache.data[ref.String()])
}

func TestSetStatus(t *testing.T) {
	setup() // server & client!
	defer teardown()
	context := &ctx.Context{GitHub: client}

	statusCache = statusMap{data: make(map[string]*statusInfo)}

	ref := newPRRef("o", "r", 1)
	ref.Repo.Quorum = 1
	newStatus := &statusInfo{
		lgtmers: []string{},
		sha:     "abcdef",
	}
	input := newStatus.NewStatus("o", ref.Repo.Quorum)

	mux.HandleFunc("/repos/o/r/statuses/abcdef", func(w http.ResponseWriter, r *http.Request) {
		v := new(github.RepoStatus)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}
		fmt.Fprint(w, `{"id":1}`)
	})

	assert.NoError(t, setStatus(
		context,
		ref,
		"abcdef",
		newStatus,
	))
	assert.Equal(t, newStatus, statusCache.data[ref.String()])
}

func TestSetStatusHTTPError(t *testing.T) {
	setup() // server & client!
	defer teardown()
	context := &ctx.Context{GitHub: client}

	statusCache = statusMap{data: make(map[string]*statusInfo)}

	ref := newPRRef("o", "r", 1)
	ref.Repo.Quorum = 1
	newStatus := &statusInfo{
		lgtmers: []string{},
		sha:     "abcdef",
	}

	mux.HandleFunc("/repos/o/r/statuses/abcdef", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "No way, Jose!", http.StatusForbidden)
	})

	assert.Error(t, setStatus(
		context,
		ref,
		"abcdef",
		newStatus,
	))
	assert.Nil(t, statusCache.data[ref.String()])
}

func TestNewEmptyStatus(t *testing.T) {
	cases := []struct {
		owner      string
		expContext string
	}{
		{"deadbeef", "deadbeef/lgtm"},
		{"jekyll", "jekyll/lgtm"},
	}
	for _, test := range cases {
		status := newEmptyStatus(test.owner)
		assert.Equal(t, test.expContext, *status.Context)
		assert.Equal(t, "failure", *status.State)
		assert.Equal(t, descriptionNoLGTMers, *status.Description)
	}
}
