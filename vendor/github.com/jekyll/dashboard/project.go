package dashboard

import (
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	defaultProjectMap map[string]*Project
	defaultProjects   = []*Project{
		newProject("jekyll", "jekyll/jekyll", "master", "jekyll"),
		newProject("jemoji", "jekyll/jemoji", "master", "jemoji"),
		newProject("mercenary", "jekyll/mercenary", "master", "mercenary"),
		newProject("jekyll-import", "jekyll/jekyll-import", "master", "jekyll-import"),
		newProject("jekyll-feed", "jekyll/jekyll-feed", "master", "jekyll-feed"),
		newProject("jekyll-seo-tag", "jekyll/jekyll-seo-tag", "master", "jekyll-seo-tag"),
		newProject("jekyll-sitemap", "jekyll/jekyll-sitemap", "master", "jekyll-sitemap"),
		newProject("jekyll-mentions", "jekyll/jekyll-mentions", "master", "jekyll-mentions"),
		newProject("jekyll-watch", "jekyll/jekyll-watch", "master", "jekyll-watch"),
		newProject("jekyll-compose", "jekyll/jekyll-compose", "master", "jekyll-compose"),
		newProject("jekyll-paginate", "jekyll/jekyll-paginate", "master", "jekyll-paginate"),
		newProject("jekyll-gist", "jekyll/jekyll-gist", "master", "jekyll-gist"),
		newProject("jekyll-coffeescript", "jekyll/jekyll-coffeescript", "master", "jekyll-coffeescript"),
		newProject("jekyll-opal", "jekyll/jekyll-opal", "master", "jekyll-opal"),
		newProject("classifier-reborn", "jekyll/classifier-reborn", "master", "classifier-reborn"),
		newProject("jekyll-sass-converter", "jekyll/jekyll-sass-converter", "master", "jekyll-sass-converter"),
		newProject("jekyll-textile-converter", "jekyll/jekyll-textile-converter", "master", "jekyll-textile-converter"),
		newProject("jekyll-redirect-from", "jekyll/jekyll-redirect-from", "master", "jekyll-redirect-from"),
		newProject("github-metadata", "jekyll/github-metadata", "master", "jekyll-github-metadata"),
		newProject("plugins website", "jekyll/plugins", "gh-pages", ""),
		newProject("jekyll docker", "jekyll/docker", "master", ""),
	}
)

func init() {
	go resetProjectsPeriodically()
}

func resetProjectsPeriodically() {
	for range time.Tick(time.Hour / 2) {
		log.Println("resetting projects' cache")
		resetProjects()
	}
}

func resetProjects() {
	for _, p := range defaultProjects {
		p.reset()
	}
}

type Project struct {
	Name    string `json:"name"`
	Nwo     string `json:"nwo"`
	Branch  string `json:"branch"`
	GemName string `json:"gem_name"`

	Gem     *RubyGem      `json:"gem"`
	Travis  *TravisReport `json:"travis"`
	GitHub  *GitHub       `json:"github"`
	fetched bool
}

func (p *Project) fetch() {
	rubyGemChan := rubygem(p.GemName)
	travisChan := travis(p.Nwo, p.Branch)
	githubChan := github(p.Nwo)

	if p.Gem == nil {
		p.Gem = <-rubyGemChan
	}

	if p.Travis == nil {
		p.Travis = <-travisChan
	}

	if p.GitHub == nil {
		p.GitHub = <-githubChan
	}

	p.fetched = true
}

func (p *Project) reset() {
	p.fetched = false
	p.Gem = nil
	p.Travis = nil
	p.GitHub = nil
}

func buildProjectMap() {
	defaultProjectMap = map[string]*Project{}
	for _, p := range defaultProjects {
		defaultProjectMap[p.Name] = p
	}
}

func newProject(name, nwo, branch, rubygem string) *Project {
	return &Project{
		Name:    name,
		Nwo:     nwo,
		Branch:  branch,
		GemName: rubygem,
	}
}

func getProject(name string) Project {
	if defaultProjectMap == nil {
		buildProjectMap()
	}

	if p, ok := defaultProjectMap[name]; ok {
		if !p.fetched {
			p.fetch()
		}
		return *p
	}
	panic(fmt.Sprintf("no project named '%s'", name))
}

func getAllProjects() []*Project {
	var wg sync.WaitGroup
	for _, p := range defaultProjects {
		wg.Add(1)
		go func(project *Project) {
			project.fetch()
			wg.Done()
		}(p)
	}
	wg.Wait()
	return defaultProjects
}

func getProjects() []*Project {
	return defaultProjects
}
