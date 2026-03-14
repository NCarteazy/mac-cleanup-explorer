.PHONY: build run test clean

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
