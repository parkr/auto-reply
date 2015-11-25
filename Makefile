build:
	godep go build
	godep go build ./cmd/unearth

deps:
	go get github.com/tools/godep

server: build
	source .env && ./auto-reply

unearth: build
	source .env && ./unearth
