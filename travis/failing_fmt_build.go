package travis

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/search"
	"github.com/parkr/githubapi/githubsearch"
)

var (
	travisAPIBaseURL      = "https://api.travis-ci.org"
	travisAPIContentType  = "application/vnd.travis-ci.2+json"
	failingFmtBuildLabels = []string{"tests", "help-wanted"}
)

type travisBuild struct {
	JobIDs []int64 `json:"job_ids"`
}

type travisJob struct {
	State  string
	Config travisJobConfig
}

type travisJobConfig struct {
	Env string
}

func FailingFmtBuildHandler(context *ctx.Context, payload interface{}) error {
	status, ok := payload.(*github.StatusEvent)
	if !ok {
		return context.NewError("FailingFmtBuildHandler: not an status event")
	}

	if *status.State != "failure" {
		return context.NewError("FailingFmtBuildHandler: not a failure status event")
	}

	if *status.Context != "continuous-integration/travis-ci/push" {
		return context.NewError("FailingFmtBuildHandler: not a continuous-integration/travis-ci/push context")
	}

	if status.Branches != nil && len(status.Branches) > 0 && *status.Branches[0].Name != "master" {
		return context.NewError("FailingFmtBuildHandler: not a travis build on the master branch")
	}

	context.SetRepo(*status.Repo.Owner.Login, *status.Repo.Name)

	buildID, err := buildIDFromTargetURL(*status.TargetURL)
	if err != nil {
		return context.NewError("FailingFmtBuildHandler: couldn't extract build ID from %q: %+v", *status.TargetURL, err)
	}
	uri := fmt.Sprintf("/repos/%s/%s/builds/%d", context.Repo.Owner, context.Repo.Name, buildID)
	resp, err := httpGetTravis(uri)
	if err != nil {
		return context.NewError("FailingFmtBuildHandler: %+v", err)
	}

	build := struct {
		Build travisBuild `json:"build"`
	}{Build: travisBuild{}}
	err = json.NewDecoder(resp.Body).Decode(&build)
	if err != nil {
		return context.NewError("FailingFmtBuildHandler: couldn't decode build json: %+v", err)
	}
	log.Printf("FailingFmtBuildHandler: %q response: %+v %+v", uri, resp, build)

	for _, jobID := range build.Build.JobIDs {
		job := struct {
			Job travisJob `json:"job"`
		}{Job: travisJob{}}
		resp, err := httpGetTravis("/jobs/" + strconv.FormatInt(jobID, 10))
		if err != nil {
			return context.NewError("FailingFmtBuildHandler: couldn't get job info from travis: %+v", err)
		}
		err = json.NewDecoder(resp.Body).Decode(&job)
		if err != nil {
			return context.NewError("FailingFmtBuildHandler: couldn't decode job json: %+v", err)
		}
		log.Printf("FailingFmtBuildHandler: job %d response: %+v %+v", jobID, resp, job)
		if job.Job.State == "failed" && job.Job.Config.Env == "TEST_SUITE=fmt" {
			// Winner! Open an issue if there isn't already one.
			query := githubsearch.IssueSearchParameters{
				Repository: &githubsearch.RepositoryName{
					Owner: context.Repo.Owner,
					Name:  context.Repo.Name,
				},
				State: githubsearch.Open,
				Scope: githubsearch.TitleScope,
				Query: "fmt is failing on master",
			}
			issues, err := search.GitHubIssues(context, query)
			if err != nil {
				return context.NewError("FailingFmtBuildHandler: couldn't run query %q: %+v", query, err)
			}
			if len(issues) > 0 {
				log.Printf("We already have an issue or issues for this failure! %s", *issues[0].HTMLURL)
			} else {
				jobHTMLURL := fmt.Sprintf("https://travis-ci.org/%s/%s/jobs/%d", context.Repo.Owner, context.Repo.Name, jobID)
				issue, _, err := context.GitHub.Issues.Create(
					context.Context(),
					context.Repo.Owner, context.Repo.Name,
					&github.IssueRequest{
						Title: github.String("fmt build is failing on master"),
						Body: github.String(fmt.Sprintf(
							"Hey @jekyll/maintainers!\n\nIt looks like the fmt build in Travis is failing again: %s :frowning_face:\n\nCould someone please fix this up? Clone down the repo, run `bundle install`, then `script/fmt` to see the failures. File a PR once you're done and say \"Fixes <this issue url>\" in the description.\n\nThanks! :revolving_hearts:",
							jobHTMLURL,
						)),
						Labels: &failingFmtBuildLabels,
					})
				if err != nil {
					return context.NewError("FailingFmtBuildHandler: failed to file an issue: %+v", err)
				}
				log.Printf("Filed issue: %s", *issue.HTMLURL)
			}
			break // you found the right job, now c'est fin
		}
	}

	return nil
}

func httpGetTravis(uri string) (*http.Response, error) {
	url := travisAPIBaseURL + uri
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("couldn't create request: %+v", err)
	}
	req.Header.Add("Accept", travisAPIContentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("couldn't send request to %s: %+v", url, err)
	}
	return resp, err
}

func buildIDFromTargetURL(targetURL string) (int64, error) {
	pieces := strings.Split(targetURL, "/")
	return strconv.ParseInt(pieces[len(pieces)-1], 10, 64)
}
