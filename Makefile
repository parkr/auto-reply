build: deps
	godep go build

deps:
	go get github.com/tools/godep

server: build
	source .env && ./auto-reply
