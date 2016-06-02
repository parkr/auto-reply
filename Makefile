all: deps build test

deps:
	go get github.com/tools/godep

updatedeps:
	godep save github.com/parkr/changelog \
	  golang.org/x/oauth2 \
	  github.com/google/go-github/github

build:
	godep go build
	godep go build ./cmd/unearth

test:
	godep go test ./...

server: build
	source .env && ./auto-reply

unearth: build
	source .env && ./unearth

clean:
	rm auto-reply unearth
