package dashboard

import (
	"log"
	"sync"
	"time"
)

var defaultProjectMap = sync.Map{}

func init() {
	for _, p := range defaultProjects {
		defaultProjectMap.Store(p.Name, p)
	}
	go resetProjectsPeriodically()
	go prefillAllProjectsFromGitHub()
}

func resetProjectsPeriodically() {
	for range time.Tick(time.Hour / 2) {
		log.Println("resetting projects' cache")
		resetProjects()
		prefillAllProjectsFromGitHub()
	}
}

func resetProjects() {
	githubGraphQLData = &githubGraphQLResults{}
	for _, p := range defaultProjects {
		p.reset()
	}
}

type Project struct {
	GlobalRelayID string `json:"id"`
	Name          string `json:"name"`
	Nwo           string `json:"nwo"`
	Branch        string `json:"branch"`
	GemName       string `json:"gem_name"`

	Gem      *RubyGem      `json:"gem"`
	Travis   *TravisReport `json:"travis"`
	GitHub   *GitHub       `json:"github"`
	AppVeyor *AppVeyor     `json:"app_veyor"`
	fetched  bool
}

func (p *Project) fetch() {
	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		p.fetchGitHubData()
		wg.Done()
	}()

	go func() {
		p.fetchRubyGemData()
		wg.Done()
	}()

	go func() {
		p.fetchTravisData()
		wg.Done()
	}()

	go func() {
		p.fetchAppVeyorData()
		wg.Done()
	}()

	wg.Wait()

	p.fetched = true
}

func (p *Project) fetchRubyGemData() {
	if p.Gem != nil {
		return
	}

	p.Gem = <-rubygem(p.GemName)
}

func (p *Project) fetchTravisData() {
	if p.Travis != nil {
		return
	}

	p.Travis = <-travis(p.Nwo, p.Branch)
}

func (p *Project) fetchGitHubData() {
	if p.GitHub != nil {
		return
	}

	p.GitHub = <-github(p.GlobalRelayID)
}

func (p *Project) fetchAppVeyorData() {
	if p.AppVeyor != nil {
		return
	}

	p.AppVeyor = <-appVeyor(p.Nwo)
}

func (p *Project) reset() {
	p.fetched = false
	p.Gem = nil
	p.Travis = nil
	p.GitHub = nil
	p.AppVeyor = nil
}

func getProject(name string) *Project {
	if p, ok := defaultProjectMap.Load(name); ok {
		proj := p.(*Project)
		if !proj.fetched {
			proj.fetch()
		}
		return proj
	}

	return nil
}

func getProjects() []*Project {
	return defaultProjects
}
