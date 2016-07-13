.PHONY: all
all: deps build test

.PHONY: deps
deps:
	go get github.com/tools/godep

.PHONY: build
build:
	godep go build -o ./bin/jekyllbot ./cmd/jekyllbot
	godep go build -o ./bin/unearth ./cmd/unearth
	godep go build -o ./bin/mark-and-sweep-stale-issues ./cmd/mark-and-sweep-stale-issues

.PHONY: test
test:
	godep go test ./...

.PHONY: server
server: build
	source .env && ./bin/auto-reply

.PHONY: unearth
unearth: build
	source .env && ./bin/unearth

.PHONY: mark-and-sweep
mark-and-sweep: build
	source .env && ./bin/mark-and-sweep-stale-issues

.PHONY: clean
clean:
	rm -r bin
