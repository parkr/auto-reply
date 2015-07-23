package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/parkr/auto-reply/Godeps/_workspace/src/github.com/google/go-github/github"
	"github.com/parkr/auto-reply/Godeps/_workspace/src/golang.org/x/oauth2"
)

const deprecatedJekyllMessage = `This repository is no longer maintained. If you're still experiencing this problem, please search for your issue on [Jekyll Talk](https://talk.jekyllrb.com/), our new community forum. If it isn't there, feel free to post to the Help category and someone will assist you. Thanks!`

var (
	client          *github.Client
	deprecatedRepos = map[string]string{
		"jekyll/jekyll-help": deprecatedJekyllMessage,
	}
)

func deprecatedReposHandler(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("error reading in body:", err)
		http.Error(w, "no req body", 400)
		return
	}

	var issue github.IssueActivityEvent
	err = json.Unmarshal(data, &issue)
	if err != nil {
		log.Println("error unmarshalling issue stuffs:", err)
		http.Error(w, "bad json", 400)
		return
	}

	if *issue.Action != "opened" {
		http.Error(w, "irrelevant issue action", 400)
		return
	}

	if msg, ok := deprecatedRepos[*issue.Repo.FullName]; ok {
		_, _, err := client.Issues.CreateComment(
			*issue.Repo.Owner.Login,
			*issue.Repo.Name,
			*issue.Issue.Number,
			&github.IssueComment{Body: &msg},
		)
		if err != nil {
			log.Println("error leaving comment:", err)
			http.Error(w, "couldnt leave comment", 500)
			return
		}
	} else {
		log.Printf("looks like '%s' repo isn't deprecated", *issue.Repo.FullName)
		http.Error(w, "non-deprecated repo", 404)
		return
	}

	w.Write([]byte(`sorry ur deprecated`))
}

func newClient() *github.Client {
	if token := os.Getenv("GITHUB_ACCESS_TOKEN"); token != "" {
		return github.NewClient(oauth2.NewClient(
			oauth2.NoContext,
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
			),
		))
	} else {
		log.Fatal("GITHUB_ACCESS_TOKEN required")
		return nil
	}
}

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "The port to serve to")
	flag.Parse()
	client = newClient()
	http.HandleFunc("/_github/repos/deprecated", deprecatedReposHandler)
	log.Printf("Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
