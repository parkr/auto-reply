all: build test

deps:
	godep save github.com/jekyll/dashboard/...

build: deps
	godep go install github.com/jekyll/dashboard/...

test: deps
	godep go test . ./cmd/... ./triage/...

