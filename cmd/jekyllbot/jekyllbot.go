// jekyllbot is the server which controls @jekyllbot on GitHub.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/parkr/auto-reply/ctx"
	"github.com/parkr/auto-reply/hooks"
	"github.com/parkr/auto-reply/jekyll"
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
	http.Handle("/_github/jekyll", jekyllOrgHandler)

	log.Printf("Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func verifyPayload(secret []byte, handler hooks.HookHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		payload, err := github.ValidatePayload(r, secret)
		if err != nil {
			log.Println("received invalid signature:", err)
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		handler.HandlePayload(w, r, payload)
	})
}
