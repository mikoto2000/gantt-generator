PROJECT := ganttgen
PKG := ./cmd/ganttgen

PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64 \
	windows/arm64

OUTPUT_DIR := dist

.PHONY: all clean build

all: build

build: clean
	@mkdir -p $(OUTPUT_DIR)
	@set -e; \
	for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; GOARCH=$${platform#*/}; \
		ext=""; [ $$GOOS = "windows" ] && ext=".exe"; \
		out="$(OUTPUT_DIR)/$(PROJECT)_$${GOOS}_$${GOARCH}$$ext"; \
		echo "Building $$out"; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build -o $$out $(PKG); \
	done

clean:
	@rm -rf $(OUTPUT_DIR)

.PHONY: fmt test

fmt:
	gofmt -w $(shell go list -f '{{.Dir}}' ./...)

test:
	go test ./...
