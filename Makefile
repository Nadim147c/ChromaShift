APP_NAME = chromashift
BIN_NAME = cshift

VERSION ?= $(shell git describe --tags)

BIN_DIR = ./bin
SRC_DIR = ./cmd
ARCHIVE_DIR = ./archive
SCRIPTS_DIR = ./scripts
COMPLETIONS_DIR = ./completions

BUILD_FLAGS = -X $(BIN_NAME)/cmd.Version=$(VERSION)

BUILD = go build -ldflags "$(BUILD_FLAGS)"

all: build

dependencies: .dependencies-stamp

.dependencies-stamp:
	@echo "Installing dependencies..."
	go mod download
	go mod verify
	go get -v

	@touch .dependencies-stamp

build:
	@echo "Building $(BIN_NAME)..."
	$(BUILD) -o $(BIN_NAME)

	@touch .build-stamp

.build-stamp: .dependencies-stamp
	@echo "Building $(BIN_NAME)..."
	$(BUILD) -o $(BIN_NAME)

	@touch .build-stamp

install:
	go install -ldflags "$(BUILD_FLAGS)"

run:
	go run main.go -- $(CMD)

test: .build-stamp
	./$(BIN_NAME) --color=always -- go test $(SRC_DIR) -failfast -v -parallel 4

test-all:
	go test $(SRC_DIR) -v -parallel 4

alias: .alias-stamp

.alias-stamp: .build-stamp
	mkdir -pv $(COMPLETIONS_DIR)
	./$(BIN_NAME) alias zsh > $(SCRIPTS_DIR)/alias.zsh
	./$(BIN_NAME) alias bash > $(SCRIPTS_DIR)/alias.bash

	@touch .alias-stamp

completion: .completion-stamp

.completion-stamp: .build-stamp
	mkdir -pv $(COMPLETIONS_DIR)
	./$(BIN_NAME) completion zsh > $(COMPLETIONS_DIR)/_$(BIN_NAME)
	./$(BIN_NAME) completion bash > $(COMPLETIONS_DIR)/$(BIN_NAME).bash
	./$(BIN_NAME) completion fish > $(COMPLETIONS_DIR)/$(BIN_NAME).fish

	@touch .completion-stamp

clean:
	rm -vrf $(BIN_DIR) $(BIN_NAME) $(ARCHIVE_DIR) $(COMPLETIONS_DIR) .*-stamp

compile:
	mkdir -vp $(BIN_DIR)

	@echo "Compiling for Unix-like OS and Platforms"
	# Linux
	GOOS=linux GOARCH=amd64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-linux-amd64 
	GOOS=linux GOARCH=arm $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-linux-arm 
	GOOS=linux GOARCH=arm64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-linux-arm64 
	
	# Darwin (macOS)
	GOOS=darwin GOARCH=amd64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-darwin-amd64 
	GOOS=darwin GOARCH=arm64 $(BUILD) -o $(BIN_DIR)/$(APP_NAME)-darwin-arm64 

archive: .alias-stamp .completion-stamp compile
	mkdir -p $(ARCHIVE_DIR)
	
	find $(BIN_DIR) -iname "$(APP_NAME)-*-*" | while read binary; do \
		basename=$$(basename $$binary); \
		printf "\n\n%s\n\n" "Archiving : $$basename"; \
		cp -vf $$binary $(BIN_DIR)/$(BIN_NAME); \
		tar -czvf "$(ARCHIVE_DIR)/$${basename}-$(VERSION).tar.gz" $(BIN_DIR)/$(BIN_NAME) $(SCRIPTS_DIR)/* $(COMPLETIONS_DIR)/* ; \
		echo "Created archive: $(ARCHIVE_DIR)/$${basename}-$(VERSION).tar.gz"; \
	done
