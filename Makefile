APP_NAME = colorize
VERSION ?= $(shell git describe --tags --abbrev=0)

BIN_DIR = ./bin
ARCHIVE_DIR = ./archive
COMPLETIONS_DIR = ./completions

BUILD = go build -ldflags "-X colorize/cmd.Version=$(VERSION)"

all: run

deps:
	go mod download
	go mod verify
	go get -v

run:
	go run main.go

APP_BIN ?= $(shell which $(APP_NAME))
APP_BIN ?= $(shell which $(BIN_DIR)/$(APP_NAME))

ifdef APP_BIN
	TEST_BIN = $(APP_BIN) --color=always -- go test
else
	TEST_BIN = go test
endif

test:
	$(TEST_BIN) ./cmd/ -vet=off -failfast -v -parallel 4

test-all:
	go test ./cmd/ -ver=off -v -parallel 4

build: deps
	$(BUILD) -o $(BIN_DIR)/$(APP_NAME)

completion: build
	mkdir -p $(COMPLETIONS_DIR)
	$(BIN_DIR)/$(APP_NAME) completion zsh > $(COMPLETIONS_DIR)/_colorize
	$(BIN_DIR)/$(APP_NAME) completion bash > $(COMPLETIONS_DIR)/colorize.bash
	$(BIN_DIR)/$(APP_NAME) completion fish > $(COMPLETIONS_DIR)/colorize.fish

clean:
	rm -rf $(BIN_DIR) $(APP_NAME) $(ARCHIVE_DIR) $(COMPLETIONS_DIR)

compile:
	mkdir -p $(BIN_DIR)
	echo "Compiling for Unix-like OS and Platforms"
	# Linux
	GOOS=linux GOARCH=amd64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 
	GOOS=linux GOARCH=arm $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-linux-arm 
	GOOS=linux GOARCH=arm64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-linux-arm64 
	
	# FreeBSD
	GOOS=freebsd GOARCH=amd64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-freebsd-amd64 
	GOOS=freebsd GOARCH=arm $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-freebsd-arm 
	GOOS=freebsd GOARCH=arm64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-freebsd-arm64 
	
	# OpenBSD
	GOOS=openbsd GOARCH=amd64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-openbsd-amd64 
	GOOS=openbsd GOARCH=arm $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-openbsd-arm 
	GOOS=openbsd GOARCH=arm64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-openbsd-arm64 
	
	# NetBSD
	GOOS=netbsd GOARCH=amd64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-netbsd-amd64 
	GOOS=netbsd GOARCH=arm $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-netbsd-arm 
	GOOS=netbsd GOARCH=arm64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-netbsd-arm64 
	
	# Darwin (macOS)
	GOOS=darwin GOARCH=amd64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-darwin-amd64 
	GOOS=darwin GOARCH=arm64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-darwin-arm64 

archive: completion compile
	mkdir -p $(ARCHIVE_DIR)
	
	find $(BIN_DIR) -iname "$(APP_NAME)-*-*" | while read binary; do \
		basename=$$(basename $$binary); \
		cp -vf $$binary $(BIN_DIR)/$(APP_NAME); \
		tar -czvf $(ARCHIVE_DIR)/$${basename}-$(VERSION).tar.gz bin/colorize scripts/* completions/* ; \
		echo "Created archive: $(ARCHIVE_DIR)/$${basename}-$(VERSION).tar.gz"; \
	done

