package dashboard

import (
	"fmt"
	"log"
)

// AppVeyor is a struct representing a given AppVeyor project & build.
type AppVeyor struct {
	Nwo     string          `json:"nwo"`
	Project appVeyorProject `json:"project"`
	Build   appVeyorBuild   `json:"build"`

	// Present if an error occurred.
	Message string `json:"message"`
}

type appVeyorProject struct {
	AccountName string `json:"accountName"`
	Name        string `json:"name"`
}

type appVeyorBuild struct {
	Status      string `json:"status"`
	BuildNumber int    `json:"buildNumber"`
	HTMLURL     string `json:"html_url"`
}

func getAppVeyor(nwo string) (*AppVeyor, error) {
	if nwo == "" {
		return nil, nil
	}

	info := &AppVeyor{Nwo: nwo}
	err := getRetry(5, fmt.Sprintf("https://ci.appveyor.com/api/projects/%s/branch/master", nwo), info)
	if err != nil {
		return nil, err
	}
	if info.Message != "" {
		return nil, fmt.Errorf("error from appveyor for %s: %s", nwo, info.Message)
	}

	info.Build.HTMLURL = fmt.Sprintf("http://ci.appveyor.com/project/%s/%s/build/%d",
		info.Project.AccountName,
		info.Project.Name,
		info.Build.BuildNumber,
	)

	return info, nil
}

func appVeyor(nwo string) chan *AppVeyor {
	appVeyorChan := make(chan *AppVeyor, 1)

	go func() {
		info, err := getAppVeyor(nwo)
		if err != nil {
			log.Printf("error fetching appveyor info for %s: %v", nwo, err)
		}
		appVeyorChan <- info
		close(appVeyorChan)
	}()

	return appVeyorChan
}
