package dashboard

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"goji.io"
	"goji.io/pat"
	"golang.org/x/net/context"
)

func reset(w http.ResponseWriter, r *http.Request) {
	resetProjects()
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"reset": true}`))
}

func show(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	name := pat.Param(ctx, "name")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(getProject(name))
}

func index(w http.ResponseWriter, r *http.Request) {
	indexTmpl.Execute(w, templateInfo{Projects: getProjects()})
}

func getBindPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8000"
}

func Listen() {
	mux := goji.NewMux()
	mux.HandleFunc(pat.Post("/reset.json"), reset)
	mux.HandleFuncC(pat.Get("/:name.json"), show)
	mux.HandleFunc(pat.Get("/"), index)

	bind := fmt.Sprintf(":%s", getBindPort())
	log.Fatal(http.ListenAndServe(bind, mux))
}
