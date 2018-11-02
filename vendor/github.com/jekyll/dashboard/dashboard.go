package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jekyll/dashboard/triage"
)

var defaultPort = 8000

func jsonResponse(w http.ResponseWriter, code int, body string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write([]byte(body))
}

func reset(w http.ResponseWriter, r *http.Request) {
	resetProjects()
	jsonResponse(w, http.StatusOK, `{"reset": "true"}`)
}

func show(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		jsonResponse(w, http.StatusBadRequest, `{"error": "missing name param"}`)
		return
	}

	proj := getProject(name)
	if proj == nil {
		jsonResponse(w, http.StatusNotFound, fmt.Sprintf(`{"error": "could not find project '%s'"}`, name))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(getProject(name))
}

func index(w http.ResponseWriter, r *http.Request) {
	indexTmpl.Execute(w, templateInfo{Projects: getProjects()})
}

func Listen(bindAddr string) error {
	http.HandleFunc("/reset.json", reset)
	http.HandleFunc("/show.json", show)
	http.Handle("/triage", triage.New(githubClient, []string{"documentation", "bug", "enhancement"}))
	http.HandleFunc("/", index)

	return http.ListenAndServe(bindAddr, nil)
}
