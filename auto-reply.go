package main

import (
	"flag"
	"log"
	"net/http"
)

const deprecatedJekyllMessage = `This repository is no longer maintained. If you're still experiencing this problem, please search for your issue on [Jekyll Talk](https://talk.jekyllrb.com/), our new community forum. If it isn't there, feel free to post to the Help category and someone will assist you. Thanks!`

var deprecatedRepos = map[string]string{
	"jekyll/jekyll-help": deprecatedJekyllMessage,
}

func deprecatedReposHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	log.Println(r.Form)
	w.Write([]byte(`hi`))
}

func main() {
	var port string
	flag.StringVar(&port, "port", "8080", "The port to serve to")
	flag.Parse()
	http.HandleFunc("/_github/repos/deprecated", deprecatedReposHandler)
	log.Printf("Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
