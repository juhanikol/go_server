.PHONY: fmt test vet vuln check cleanbuild build clean

BINARY_NAME=example_server
BUILD_DIR=bin

fmt:
	go fmt ./...

test:
	go test ./...

vet:
	go vet ./...

vuln:
	govulncheck ./...

check: fmt test vet vuln

cleanbuild: clean build

build:
	@echo "Building app"
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/example_server
	@echo "Ready! Binary in folder: $(BUILD_DIR)/$(BINARY_NAME)"

clean:
	@echo "Cleaning old builds"
	rm -rf $(BUILD_DIR)
	go clean -cache