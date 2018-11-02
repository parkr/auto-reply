all: build test

deps:
	dep ensure

build: deps
	go install github.com/jekyll/dashboard/...

test: deps
	go test . ./cmd/... ./triage/...

server: build
	dashboard -http=localhost:8000
