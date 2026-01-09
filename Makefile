VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS = -X github.com/donghun/wm/internal/version.Version=$(VERSION) \
          -X github.com/donghun/wm/internal/version.GitCommit=$(COMMIT) \
          -X github.com/donghun/wm/internal/version.BuildDate=$(DATE)

.PHONY: build test clean install

build:
	go build -ldflags "$(LDFLAGS)" -o wm .

test:
	go test ./... -v

clean:
	rm -f wm

install:
	go install -ldflags "$(LDFLAGS)" .

# Run all tests including integration
test-all:
	go test ./... -v -count=1
