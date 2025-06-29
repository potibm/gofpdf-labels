# Simple Makefile for gofpdf-labels

APP_NAME = gofpdf-labels

# Default target: run all checks
all: tidy fmt vet lint test

# Run tests
test:
	go test ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Format code
fmt:
	go fmt ./...

# Run static analysis
vet:
	go vet ./...

# Run linter (needs golangci-lint installed)
lint:
	golangci-lint run

# Download modules & tidy up
tidy:
	go mod tidy

lint-actions:
	actionlint

prettier:
	prettier --write "**/*.yaml" "**/*.md"

# Run all linters together
lint-all: lint-go lint-actions prettier

.PHONY:lint lint-go lint-actions prettier lint-all test tidy coverage
