.DEFAULT_GOAL := build

fmt:
	go fmt ./...
vet: fmt
	go vet ./...
build: vet
	go build -o bin/service main.go
run: build
	bin/service
clean:
	if exist bin\ rmdir /s /q bin
.PHONY: fmt vet build run clean