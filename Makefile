all: deps build test

deps:
	go get github.com/tools/godep

build:
	godep go build
	godep go build ./cmd/unearth

test:
	godep go test ./...

server: build
	source .env && ./auto-reply

unearth: build
	source .env && ./unearth
