// jekyllbot is the server which controls @jekyllbot on GitHub.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/parkr/auto-reply/ctx"
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
