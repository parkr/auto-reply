package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/autopull"
	"github.com/parkr/auto-reply/comments"
	"github.com/parkr/auto-reply/common"
	"github.com/parkr/auto-reply/deprecate"
	"github.com/parkr/auto-reply/labeler"
	"github.com/parkr/auto-reply/messages"
)

var (
	client *github.Client

	deprecatedRepos = []deprecate.RepoDeprecation{
		deprecate.RepoDeprecation{
			Nwo:     "jekyll/jekyll-help",
			Message: `This repository is no longer maintained. If you're still experiencing this problem, please search for your issue on [Jekyll Talk](https://talk.jekyllrb.com/), our new community forum. If it isn't there, feel free to post to the Help category and someone will assist you. Thanks!`,
		},
	}
)

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "The port to serve to")
	flag.Parse()
	client = common.NewClient()

	http.HandleFunc("/_ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("ok\n"))
	}))

	deprecationHandler := deprecate.NewHandler(client, deprecatedRepos)
	http.HandleFunc("/_github/repos/deprecated", verifyPayload(
		getSecret("DEPRECATE"),
		deprecationHandler,
	))

	autoPullHandler := autopull.NewHandler(client, []string{"jekyll/jekyll"})
	http.HandleFunc("/_github/repos/autopull", verifyPayload(
		getSecret("AUTOPULL"),
		autoPullHandler,
	))

	commentsHandler := comments.NewHandler(client,
		[]comments.CommentHandler{
			comments.HandlerPendingFeedbackLabel,
		},
		[]comments.CommentHandler{
			comments.HandlerPendingFeedbackLabel,
			comments.HandlerMergeAndLabel,
		},
	)
	http.HandleFunc("/_github/repos/comments", verifyPayload(
		getSecret("COMMENTS"),
		commentsHandler,
	))

	labelerHandler := labeler.NewHandler(client,
		[]labeler.PushHandler{},
		[]labeler.PullRequestHandler{
			labeler.PendingRebasePRLabeler,
		},
	)
	http.HandleFunc("/_github/repos/labeler", verifyPayload(
		getSecret("LABELER"),
		labelerHandler,
	))

	log.Printf("Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getSecret(suffix string) []byte {
	return []byte(os.Getenv("GH_SECRET_" + suffix))
}

func verifyPayload(secret []byte, handler messages.PayloadHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		payload, err := messages.ValidatedPayload(r, secret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		handler.HandlePayload(w, r, payload)
	})
}
