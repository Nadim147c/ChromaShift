GO       ?= go
BIN_NAME ?= cshift
VERSION  ?= $(shell git describe --tags)
PREFIX   ?= /usr/local/
TOOL_MOD ?= -modfile tool.go.mod
TOOL     ?= $(GO) tool $(TOOL_MOD)

BIN_FILE            = $(shell realpath -m "$(PREFIX)/bin/$(BIN_NAME)")
LICENSE_FILE        = $(shell realpath -m "$(PREFIX)/share/licenses/$(BIN_NAME)/LICENSE")
BASH_COMPLETION_DIR = $(shell realpath -m "$(PREFIX)/share/bash-completion/completions")
ZSH_COMPLETION_DIR  = $(shell realpath -m "$(PREFIX)/share/zsh/site-functions")
FISH_COMPLETION_DIR = $(shell realpath -m "$(PREFIX)/share/fish/vendor_completions.d")

-include Makefile.local

.PHONY: build install test docs-dev

build:
	$(GO) build -trimpath -ldflags '-X main.Version=$(VERSION)' -o $(BIN_NAME)

install:
	install -Dm755 $(BIN_NAME) "$(BIN_FILE)"
	install -Dm644 LICENSE "$(LICENSE_FILE)"
	$(BIN_NAME) _carapace bash | install -Dm644 /dev/stdin "$(BASH_COMPLETION_DIR)/$(BIN_NAME)"
	$(BIN_NAME) _carapace zsh  | install -Dm644 /dev/stdin "$(ZSH_COMPLETION_DIR)/_$(BIN_NAME)"
	$(BIN_NAME) _carapace fish | install -Dm644 /dev/stdin "$(FISH_COMPLETION_DIR)/$(BIN_NAME).fish"

tools-install:
	$(GO) get $(TOOL_MOD) -tool github.com/mgechev/revive@latest
	$(GO) get $(TOOL_MOD) -tool github.com/segmentio/golines@latest
	$(GO) get $(TOOL_MOD) -tool mvdan.cc/gofumpt@latest
	$(GO) mod tidy $(TOOL_MOD)

format:
	find -iname '*.go' -print0 | xargs -0 $(TOOL) golines --max-len 80 -w
	find -iname '*.go' -print0 | xargs -0 $(TOOL) gofumpt -w

test:
	$(GO) test -v ./...
	$(TOOL) revive -config revive.toml -formatter friendly ./...

docs-dev:
	bun run dev

