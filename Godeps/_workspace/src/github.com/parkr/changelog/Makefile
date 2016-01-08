all: build test run

dist:
	mkdir -p dist

build: dist
	go build
	go build -o dist/changelogger ./changelogger

test:
	go test ./...

run: build
	dist/changelogger -h || true
	dist/changelogger -out=History.markdown
