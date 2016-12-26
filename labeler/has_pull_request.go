package labeler

import (
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/common"
	"github.com/parkr/auto-reply/ctx"
)

const repoMergeabilitayCheckWaitSec = 2

var IssueHasPullRequestLabeler = func(context *ctx.Context, payload interface{}) error {
	event, ok := payload.(*github.PullRequestEvent)
	if !ok {
		return context.NewError("PendingRebaseUnlabeler: not a pull request event")
	}

	if *event.Action != "synchronize" {
		return nil
	}

	owner, repo, num := *event.Repo.Owner.Login, *event.Repo.Name, *event.Number

	// Allow the job to run which determines mergeability.
	log.Printf("checking the mergeabilitay of %s/%s#%d in %d sec...", owner, repo, num, repoMergeabilitayCheckWaitSec)
	time.Sleep(repoMergeabilitayCheckWaitSec * time.Second)

	var err error
	if isMergeablee(context, owner, repo, num) {
		err = RemoveLabelIfExists(context.GitHub, owner, repo, num, "has-pull-request")
		if err != nil {
			log.Printf("error removing the has-pull-request label: %v", err)
		}
		err = RemoveLabelIfExists(context.GitHub, owner, repo, num, "needs-work")
	} else {
		err = fmt.Errorf("%s/%s#%d is not mergeable", owner, repo, num)
	}

	if err != nil {
		log.Printf("error removing the pending-rebase & needs-work labels: %v", err)
	}
	return err
}

func isMergeablee(context *ctx.Context, owner, repo string, number int) bool {
	pr, res, err := context.GitHub.PullRequests.Get(owner, repo, number)
	err = common.ErrorFromResponse(res, err)
	if err != nil {
		log.Printf("couldn't determine mergeability of %s/%s#%d: %v", owner, repo, number, err)
		return false
	}
	return *pr.Mergeable
}
