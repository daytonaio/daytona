# Default values
VERSION ?= v0.0.0-dev

ifeq ($(OS),Windows_NT)
    DAYTONA_CONFIG_DIR ?= $(APPDATA)/daytona
else
    DAYTONA_CONFIG_DIR ?= $(HOME)/.config/daytona
endif

OUTPUT_DIR ?= $(DAYTONA_CONFIG_DIR)/server/binaries/$(VERSION)

# Supported OSs and ARCHs
OS_LIST := darwin linux windows
ARCH_LIST := amd64 arm64

# Infer ARCH if not provided
ifeq ($(ARCH),)
    UNAME_M := $(shell uname -m)
    ifeq ($(findstring $(UNAME_M),x86_64 AMD64),$(UNAME_M))
        ARCH := amd64
    else ifeq ($(findstring $(UNAME_M),arm64 ARM64 aarch64),$(UNAME_M))
        ARCH := arm64
    else
        $(error Unable to infer ARCH from $(UNAME_M), please specify manually)
    endif
endif

# Default target
all: build-all

# Build for all OS and ARCH combinations
build-all:
		@for os in $(OS_LIST); do \
        for arch in $(ARCH_LIST); do \
            $(MAKE) build OS=$$os ARCH=$$arch; \
        done; \
    done

# Build for a specific OS and ARCH
build:
		@if [ -z "$(OS)" ]; then \
        echo "Error: OS must be specified."; \
        echo "Usage: make build OS=<os> [ARCH=<arch>]"; \
        exit 1; \
    fi
		@echo "Building for $(OS)-$(ARCH)"
		@mkdir -p $(OUTPUT_DIR)
		@GOOS=$(OS) GOARCH=$(ARCH) go build -o $(OUTPUT_DIR)/daytona-$(OS)-$(ARCH)$(if $(filter windows,$(OS)),.exe) cmd/daytona/main.go

# Clean build artifacts
clean:
		@rm -rf $(OUTPUT_DIR)

# Help target
help:
	@sh -c '\
		echo "Available targets:"; \
		echo "  make              : Build for all supported OS and ARCH combinations"; \
		echo "  make build OS=<os> [ARCH=<arch>] : Build for a specific OS and ARCH"; \
		echo "  make clean        : Remove build artifacts"; \
		echo ""; \
		echo "Supported OS  : $(OS_LIST)"; \
		echo "Supported ARCH: $(ARCH_LIST)"; \
		echo ""; \
		echo "Environment variables:"; \
		echo "  VERSION           : Set the version (default: v0.0.0-dev)"; \
		echo "  DAYTONA_CONFIG_DIR: Set the config directory (default: ~/.config/daytona)"; \
		echo "  OUTPUT_DIR        : Override the output directory"; \
		echo ""; \
		echo "Note: If ARCH is not specified, it will be inferred from the current machine."'

.PHONY: all build-all build clean help