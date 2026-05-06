.PHONY: dev build test vet tidy

dev:
	./scripts/dev.sh

build:
	CGO_ENABLED=1 go build -tags vips -o diegonoticias ./cmd/server

test:
	go test ./...

vet:
	go vet ./...

tidy:
	go mod tidy

