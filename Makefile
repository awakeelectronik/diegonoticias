.PHONY: dev build test vet tidy

dev:
	./scripts/dev.sh

build:
	go build -o diegonoticias ./cmd/server

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

