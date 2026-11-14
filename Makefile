# Test commands. Integration tests are behind the 'integration' build tag and their names start with 'TestIntegration'

GO ?= go

.PHONY: test test-integration test-all cover

## test: runs the unit tests
test:
	$(GO) test ./...

## test-integration: runs only the integration tests
test-integration:
	$(GO) test -tags integration -run '^TestIntegration' ./...

## test-all: runs all tests with the race detector
test-all:
	$(GO) test -race -tags integration ./...

## cover: runs all tests and prints the coverage of each function (also writes 'coverage.out')
cover:
	$(GO) test -tags integration -coverpkg=./... -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out
