.PHONY: setup build run test clean

setup:
	@command -v go >/dev/null 2>&1 || { echo "Installing Go via Homebrew..."; command -v brew >/dev/null 2>&1 || { echo "Error: Homebrew is required. Install from https://brew.sh"; exit 1; }; brew install go; }
	@echo "Installing dependencies..."
	go mod download
	@echo "Building..."
	go build -o mac-cleanup-explorer .
	@echo ""
	@echo "Ready! Run with:"
	@echo "  make run       # scan full filesystem"
	@echo "  make run-home  # scan home directory only"

build:
	go build -o mac-cleanup-explorer .

run: build
	./mac-cleanup-explorer

run-home: build
	./mac-cleanup-explorer -path $$HOME

test:
	go test ./... -v

test-race:
	go test ./... -v -race

clean:
	rm -f mac-cleanup-explorer

lint:
	go vet ./...
