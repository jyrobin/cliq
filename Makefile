BINARY := cliq
BUILD_DIR := build
INSTALL_DIR := $(HOME)/go/bin

# Build flags
LDFLAGS := -s -w
GO := go

.PHONY: all build clean test install uninstall fmt vet

all: build

build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) ./cmd/cliq

clean:
	rm -rf $(BUILD_DIR)
	$(GO) clean

test:
	$(GO) test -v ./...

install: build
	@mkdir -p $(INSTALL_DIR)
	cp $(BUILD_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "Installed $(BINARY) to $(INSTALL_DIR)"

uninstall:
	rm -f $(INSTALL_DIR)/$(BINARY)
	@echo "Uninstalled $(BINARY) from $(INSTALL_DIR)"

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

# Development helpers
run: build
	./$(BUILD_DIR)/$(BINARY) $(ARGS)

deps:
	$(GO) mod download
	$(GO) mod tidy
