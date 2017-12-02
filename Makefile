BINARIES = bin/check-for-outdated-dependencies \
    bin/jekyllbot \
    bin/mark-and-sweep-stale-issues \
    bin/nudge-maintainers-to-release \
    bin/unearth \
    bin/unify-labels

.PHONY: all
all: deps build test

.PHONY: deps
deps:
	go get github.com/tools/godep

.PHONY: $(BINARIES)
$(BINARIES): deps clean
	godep go build -o ./$@ ./$(patsubst bin/%,cmd/%,$@)

.PHONY: build
build: clean $(BINARIES)
	ls -lh bin/

.PHONY: test
test:
	godep go test github.com/parkr/auto-reply/...

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
	rm -rf bin
