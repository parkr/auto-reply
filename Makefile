ROOT_PKG=github.com/parkr/auto-reply
BINARIES = bin/check-for-outdated-dependencies \
    bin/jekyllbot \
    bin/mark-and-sweep-stale-issues \
    bin/nudge-maintainers-to-release \
    bin/unearth \
    bin/unify-labels

.PHONY: all
all: deps fmt build test

.PHONY: deps
deps:
	go get github.com/tools/godep

.PHONY: fmt
fmt:
	git ls-files | grep -v '^vendor' | grep '\.go$$' | xargs gofmt -s -l -w | sed -e 's/^/Fixed /'
	go list $(ROOT_PKG)/... | xargs go fix
	go list $(ROOT_PKG)/... | xargs go vet

.PHONY: $(BINARIES)
$(BINARIES): deps clean
	godep go build -o ./$@ ./$(patsubst bin/%,cmd/%,$@)

.PHONY: build
build: clean $(BINARIES)
	ls -lh bin/

.PHONY: test
test: deps
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
