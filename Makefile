GO         ?= go
GOFMT      ?= $(GO)fmt

BIN_NAME   ?= knock
BIN_DIR    ?= $(shell pwd)/build

VERSION    ?= $(shell cat VERSION)
REVERSION  ?=$(shell git log -1 --pretty="%H")
BRANCH     ?=$(shell git rev-parse --abbrev-ref HEAD)
TIME       ?=$(shell date +%Y-%m-%dT%H:%M:%S%z)


default: fmt style build

fmt:
	@echo ">> format code style"
	$(GOFMT) -w $$(find . -path ./vendor -prune -o -name '*.go' -print) 

style:
	@echo ">> checking code style"
	! $(GOFMT) -d $$(find . -path ./vendor -prune -o -name '*.go' -print) | grep '^'

build:
	@echo ">> building binaries"
	$(GO) build -o build/$(BIN_NAME) -ldflags  '-X "main.Version=$(VERSION)" -X  "main.BuildRevision=$(REVERSION)" -X  "main.BuildBranch=$(BRANCH)" -X "main.BuildTime=$(TIME)" '

linux:
	@echo ">> building binaries"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o build/$(BIN_NAME)-linux -ldflags  '-X "main.Version=$(VERSION)" -X  "main.BuildRevision=$(REVERSION)" -X  "main.BuildBranch=$(BRANCH)" -X "main.BuildTime=$(TIME)"'

all:  fmt style linux

clean:
	rm -rf $(BIN_DIR)

.PHONY: all fmt style build linux
