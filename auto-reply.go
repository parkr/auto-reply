package main

import (
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
	w.Write([]byte{`hi`})
}

func main() {
	http.HandleFunc("/_github/repos/deprecated", deprecatedReposHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
