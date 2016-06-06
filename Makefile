all: deps build test

deps:
	go get github.com/tools/godep

build:
	godep go build
	godep go build ./cmd/unearth
	godep go build ./cmd/mark-and-sweep-stale-issues

test:
	godep go test ./...

server: build
	source .env && ./auto-reply

unearth: build
	source .env && ./unearth

mark-and-sweep: build
	source .env && ./mark-and-sweep-stale-issues

clean:
	rm auto-reply unearth
