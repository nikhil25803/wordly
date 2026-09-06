BINARY := wordly
BIN_DIR := bin

.PHONY: build run test seed clean check release snapshot

build:
	go build -o $(BIN_DIR)/$(BINARY) ./cmd/wordly

run:
	go run ./cmd/wordly

test:
	go test ./...

seed:
	./scripts/seed-words.sh

clean:
	rm -rf $(BIN_DIR)
	rm -rf dist

check:
	goreleaser check

release:
	goreleaser release --clean

snapshot:
	goreleaser release --snapshot --clean
