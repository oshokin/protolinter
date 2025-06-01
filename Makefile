export GO111MODULE=on

APP:=protolinter
OS:=$(shell go env GOOS)
ARCH:=$(shell go env GOARCH)
LOCAL_BIN:=$(CURDIR)/bin
GOLANGCI_BIN:=$(LOCAL_BIN)/golangci-lint
GOLANGCI_TAG:=1.64.8
GOLANGCI_CONFIG:=.golangci.yaml
GOLANGCI_STRICT_CONFIG:=.golangci-strict.yaml

GOLANGCI_BIN_VERSION := $(shell $(GOLANGCI_BIN) --version 2> /dev/null | sed -E 's/.* version v(.*) built .* from .*/\1/g')
ifneq ($(GOLANGCI_BIN_VERSION),$(GOLANGCI_TAG))
GOLANGCI_BIN:=
endif

default: help

.PHONY: install-lint
install-lint:
ifeq ($(wildcard $(GOLANGCI_BIN)),)
	$(info Downloading golangci-lint v$(GOLANGCI_TAG))
	@mkdir -p $(LOCAL_BIN)
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v$(GOLANGCI_TAG)
GOLANGCI_BIN:=$(LOCAL_BIN)/golangci-lint
endif

.PHONY: lint
lint: install-lint
	$(info Running lint in normal mode...)
	$(GOLANGCI_BIN) run --new-from-rev=origin/master --config=$(GOLANGCI_CONFIG) ./...

.PHONY: lint-strict
lint-strict: install-lint
	$(info Running lint in strict mode...)
	$(GOLANGCI_BIN) run --new-from-rev=origin/master --config=$(GOLANGCI_STRICT_CONFIG) ./...

.PHONY: lint-full
lint-full: install-lint
	$(info Running lint in normal mode...)
	$(GOLANGCI_BIN) run --config=$(GOLANGCI_CONFIG) ./...

.PHONY: lint-strict-full
lint-strict-full: install-lint
	$(info Running lint in strict mode...)
	$(GOLANGCI_BIN) run --config=$(GOLANGCI_STRICT_CONFIG) ./...

.PHONY: test
test:
	@go test -v ./...

.PHONY: build
build:
	$(info Building $(APP) for $(OS)/$(ARCH))
	@mkdir -p $(LOCAL_BIN)
	@GOOS=$(OS) GOARCH=$(ARCH) go build -o $(LOCAL_BIN)/$(APP) main.go

.PHONY: run
run:
	@mkdir -p $(LOCAL_BIN)
	@$(LOCAL_BIN)/$(APP)

.PHONY: clean
clean:
	@mkdir -p $(LOCAL_BIN)
	@rm -rf $(LOCAL_BIN)/$(APP)

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  help                    Show this help message"
	@echo "  install-lint            Download and install golangci-lint to $(LOCAL_BIN) directory if it's not already installed"
	@echo "  lint                    Run golangci-lint with normal checks and compare changes against master branch"
	@echo "  lint-strict             Same as 'lint', but with more strict checks"
	@echo "  lint-full               Run golangci-lint with normal checks for all files in the repository"
	@echo "  lint-strict-full        Same as 'lint-full', but with more strict checks"
	@echo "  test                    Run unit tests"
	@echo "  build                   Build the $(APP) binary for $(OS)/$(ARCH)"
	@echo "  run                     Run the $(APP) binary"
	@echo "  clean                   Remove the $(APP) binary"
