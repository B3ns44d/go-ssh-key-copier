BINARY_NAME=ssh-key-copier
CMD_PATH=./cmd/ssh-key-copier
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS = -ldflags="-X main.version=${VERSION} -s -w" # -s -w for smaller binaries

# Default target
all: build

build: fmt
	@echo "Building ${BINARY_NAME} version ${VERSION}..."
	@go build ${LDFLAGS} -o ${BINARY_NAME} ${CMD_PATH}

install: fmt
	@echo "Installing ${BINARY_NAME} to $(go env GOPATH)/bin..."
	@go install ${LDFLAGS} ${CMD_PATH}

fmt:
	@echo "Formatting code..."
	@go fmt ./...

clean:
	@echo "Cleaning..."
	@rm -f ${BINARY_NAME}

test:
	@echo "Running tests (if any)..."
	@go test ./...

.PHONY: all build install fmt clean test