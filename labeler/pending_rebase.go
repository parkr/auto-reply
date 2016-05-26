package labeler

import (
	"fmt"
	"log"
	"time"

	"github.com/google/go-github/github"
)

const repoMergeabilityCheckWaitSec = 2

var PendingRebasePRLabeler = func(client *github.Client, event github.PullRequestEvent) error {
	if *event.Action != "synchronize" {
		return nil
	}

	owner, repo, num := *event.Repo.Owner.Login, *event.Repo.Name, *event.Number

	// Allow the job to run which determines mergeability.
	log.Printf("checking the mergeability of %s/%s#%d in %d sec...", owner, repo, num, repoMergeabilityCheckWaitSec)
	time.Sleep(repoMergeabilityCheckWaitSec * time.Second)

	var err error
	if isMergeable(client, owner, repo, num) {
		err = RemoveLabel(client, owner, repo, num, "pending-rebase")
	} else {
		err = fmt.Errorf("%s/%s#%d is not mergeable", owner, repo, num)
	}

	if err != nil {
		log.Printf("error removing the pending-rebase label: %v", err)
	}
	return err
}

func isMergeable(client *github.Client, owner, repo string, number int) bool {
	pr, _, err := client.PullRequests.Get(owner, repo, number)
	if err != nil {
		log.Printf("couldn't determine mergeability of %s/%s#%d: %v", owner, repo, number, err)
	}
	return *pr.Mergeable
}
