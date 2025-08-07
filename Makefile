# Terraform Provider for Uptime Monitor
# Private provider for uptime monitoring service

VERSION ?= dev
BINARY = terraform-provider-uptime
PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

.PHONY: build
build:
	go build -o $(BINARY) .

.PHONY: install
install: build
	@OS=$$(go env GOOS); \
	ARCH=$$(go env GOARCH); \
	mkdir -p ~/.terraform.d/plugins/localhost/uptime/uptime/$(VERSION)/$${OS}_$${ARCH}/; \
	cp $(BINARY) ~/.terraform.d/plugins/localhost/uptime/uptime/$(VERSION)/$${OS}_$${ARCH}/

.PHONY: test
test:
	go test -v ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: clean
clean:
	rm -f $(BINARY)
	rm -rf dist/

.PHONY: dist
dist: clean
	mkdir -p dist
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		echo "Building for $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build -o dist/$(BINARY)_$(VERSION)_$${GOOS}_$${GOARCH} .; \
	done

.PHONY: docs
docs:
	go generate ./...

.PHONY: testacc
testacc:
	TF_ACC=1 go test -v -timeout 120m ./...

.PHONY: all
all: fmt test build

.DEFAULT_GOAL := build