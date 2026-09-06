.PHONY: build-wordly run-wordly

build-wordly:
	go build -o bin/wordly cmd/wordly/main.go

run-wordly:
	go run cmd/wordly/main.go
