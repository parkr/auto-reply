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

var PendingRebaseNeedsWorkPRUnlabeler = func(context *ctx.Context, event github.PullRequestEvent) error {
	if *event.Action != "synchronize" {
		return nil
	}

	owner, repo, num := *event.Repo.Owner.Login, *event.Repo.Name, *event.Number

	// Allow the job to run which determines mergeability.
	log.Printf("checking the mergeability of %s/%s#%d in %d sec...", owner, repo, num, repoMergeabilityCheckWaitSec)
	time.Sleep(repoMergeabilityCheckWaitSec * time.Second)

	var err error
	if isMergeable(context, owner, repo, num) {
		err = RemoveLabelIfExists(context.GitHub, owner, repo, num, "pending-rebase")
		if err != nil {
			log.Printf("error removing the pending-rebase label: %v", err)
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

func isMergeable(context *ctx.Context, owner, repo string, number int) bool {
	pr, res, err := context.GitHub.PullRequests.Get(owner, repo, number)
	err = common.ErrorFromResponse(res, err)
	if err != nil {
		log.Printf("couldn't determine mergeability of %s/%s#%d: %v", owner, repo, number, err)
		return false
	}
	return *pr.Mergeable
}
