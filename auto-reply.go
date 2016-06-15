package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/parkr/auto-reply/autopull"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/hooks"
	"github.com/parkr/auto-reply/jekyll"
	"github.com/parkr/auto-reply/labeler"
	"github.com/parkr/auto-reply/messages"
)

var context *ctx.Context

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "The port to serve to")
	flag.Parse()
	context = ctx.NewDefaultContext()

	http.HandleFunc("/_ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("ok\n"))
	}))

	jekyllOrgHandler := jekyll.NewJekyllOrgHandler(context)
	http.HandleFunc("/_github/jekyll", verifyPayload(
		getSecret("JEKYLL"),
		jekyllOrgHandler,
	))

	autoPullHandler := autopull.NewHandler(context, []string{"jekyll/jekyll"})
	http.HandleFunc("/_github/repos/autopull", verifyPayload(
		getSecret("AUTOPULL"),
		autoPullHandler,
	))

	labelerHandler := labeler.NewHandler(context,
		[]labeler.PushHandler{},
		[]labeler.PullRequestHandler{
			labeler.PendingRebaseNeedsWorkPRUnlabeler,
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

func verifyPayload(secret []byte, handler hooks.HookHandler) http.HandlerFunc {
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
