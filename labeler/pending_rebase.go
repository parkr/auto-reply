package labeler

import (
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/common"
	"github.com/parkr/auto-reply/ctx"
)

const repoMergeabilityCheckWaitSec = 2

var PendingRebaseNeedsWorkPRUnlabeler = func(context *ctx.Context, payload interface{}) error {
	event, ok := payload.(*github.PullRequestEvent)
	if !ok {
		return context.NewError("PendingRebaseUnlabeler: not a pull request event")
	}

	if *event.Action != "synchronize" {
		return nil
	}

	owner, repo, num := *event.Repo.Owner.Login, *event.Repo.Name, *event.Number

	// Allow the job to run which determines mergeability.
	log.Printf("checking the mergeability of %s/%s#%d in %d sec...", owner, repo, num, repoMergeabilityCheckWaitSec)
	time.Sleep(repoMergeabilityCheckWaitSec * time.Second)

	var err error
	if isMergeable(context, owner, repo, num) {
		err = RemoveLabelIfExists(context, owner, repo, num, "pending-rebase")
		if err != nil {
			log.Printf("error removing the pending-rebase label: %v", err)
		}
		err = RemoveLabelIfExists(context, owner, repo, num, "needs-work")
	} else {
		err = fmt.Errorf("%s/%s#%d is not mergeable", owner, repo, num)
	}

	if err != nil {
		log.Printf("error removing the pending-rebase & needs-work labels: %v", err)
	}
	return err
}

func isMergeable(context *ctx.Context, owner, repo string, number int) bool {
	if context == nil {
		panic("context cannot be nil!")
	}

	pr, res, err := context.GitHub.PullRequests.Get(context.Context(), owner, repo, number)
	err = common.ErrorFromResponse(res, err)
	if err != nil {
		log.Printf("couldn't determine mergeability of %s/%s#%d: %v", owner, repo, number, err)
		return false
	}

	if pr == nil {
		log.Printf("isMergeable: %s/%s#%d appears not to exist", owner, repo, number)
		return false
	}

	if pr.Mergeable == nil {
		log.Printf("isMergeable: %s/%s#%d mergability is not populated in the API response", owner, repo, number)
		return false
	}

	return *pr.Mergeable
}
