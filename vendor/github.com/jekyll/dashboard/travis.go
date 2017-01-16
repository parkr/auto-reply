package dashboard

import (
	"fmt"
	"log"
)

type TravisReport struct {
	Nwo    string       `json:"nwo"`
	Branch TravisBranch `json:"branch"`
}

type TravisBranch struct {
	Id    int    `json:"id"`
	State string `json:"state"`
}

func travis(nwo, branch string) chan *TravisReport {
	travisChan := make(chan *TravisReport, 1)

	go func() {
		if branch == "" {
			travisChan <- nil
			return
		}

		var info TravisReport
		info.Nwo = nwo

		err := getRetry(5, fmt.Sprintf("https://api.travis-ci.org/repos/%s/branches/%s", nwo, branch), &info)
		if err != nil {
			travisChan <- nil
			log.Printf("==> error fetching travis info for %s/%s: %v", nwo, branch, err)
		}

		travisChan <- &info
		close(travisChan)
	}()

	return travisChan
}
