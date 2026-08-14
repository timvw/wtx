BINARY  := wtx
PKG     := github.com/timvw/wtx

# Version stamped into the binary. `git describe` falls back to a bare commit
# hash via --always, so this works in a repository with no tags yet; the outer
# fallback covers a source tree that is not a git checkout at all.
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X $(PKG)/cmd.version=$(VERSION)

.DEFAULT_GOAL := build

.PHONY: build test vet fmt clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

clean:
	rm -f $(BINARY)
