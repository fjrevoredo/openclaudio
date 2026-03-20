GO ?= /usr/local/go/bin/go
GOCACHE ?= /tmp/openclaudio-gocache
GOPATH ?= /tmp/openclaudio-gopath

.PHONY: frontend-build build test run

frontend-build:
	npm run build

build: frontend-build
	mkdir -p $(GOCACHE) $(GOPATH)
	PATH=/usr/local/go/bin:$$PATH GOCACHE=$(GOCACHE) GOPATH=$(GOPATH) $(GO) build -o openclaudio ./cmd/openclaudio

test:
	mkdir -p $(GOCACHE) $(GOPATH)
	PATH=/usr/local/go/bin:$$PATH GOCACHE=$(GOCACHE) GOPATH=$(GOPATH) $(GO) test ./...

run: frontend-build
	mkdir -p $(GOCACHE) $(GOPATH)
	PATH=/usr/local/go/bin:$$PATH GOCACHE=$(GOCACHE) GOPATH=$(GOPATH) $(GO) run ./cmd/openclaudio
