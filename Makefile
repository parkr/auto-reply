.PHONY: all
all: deps build test

.PHONY: deps
deps:
	go get github.com/tools/godep

.PHONY: build
build:
	godep go build
	godep go build ./cmd/unearth
	godep go build ./cmd/mark-and-sweep-stale-issues

.PHONY: test
test:
	godep go test ./...

.PHONY: server
server: build
	source .env && ./auto-reply

.PHONY: unearth
unearth: build
	source .env && ./unearth

.PHONY: mark-and-sweep
mark-and-sweep: build
	source .env && ./mark-and-sweep-stale-issues

.PHONY: clean
clean:
	rm auto-reply unearth mark-and-sweep-stale-issues
