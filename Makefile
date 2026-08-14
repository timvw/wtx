BINARY  := wtx
PKG     := github.com/timvw/wtx

# Version stamped into the binary. `git describe` falls back to a bare commit
# hash via --always, so this works in a repository with no tags yet; the outer
# fallback covers a source tree that is not a git checkout at all.
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X $(PKG)/cmd.version=$(VERSION)

.DEFAULT_GOAL := build

.PHONY: build test vet fmt fmt-check lint clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l -w .

# What CI checks: report unformatted files and fail, never rewrite them. `fmt`
# is the fix; this is the gate.
fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "not gofmt-formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

# Uses .golangci.yml at the repository root, the same file CI's Lint job reads.
lint:
	golangci-lint run

clean:
	rm -f $(BINARY)
	rm -rf dist
