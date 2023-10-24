# ==================================================================================== #
#
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run: run application
.PHONY: run
run:
	go run vd.go

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: tidy and vendor dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...

## tests: run tests for monitoring
.PHONY: tests
tests:
	@echo 'Running tests...'
	go test -race -vet=off ./...
	go test -coverpkg=./... -coverprofile=/tmp/coverage.out ./...
	go tool cover -func /tmp/coverage.out | tail -1

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

# ==================================================================================== #
# BUILD
# ==================================================================================== #

## build: build the application
.PHONY: build
build:
	@echo 'Building virtual device...'
	go build .

## build/vd/x86-64: build binary for x86-64
.PHONY: build/x86-64
build/x86-64:
	@echo 'Building cmd/vd for x86_64...'
	go build -o=./bin/x86_64/vd .

## build/vd/arm: build binary for arm
.PHONY: build/arm
build/arm:
	@echo 'Building cmd/vd for arm...'
	GOOS=linux GOARCH=arm go build -o=./bin/arm/vd .

## build/gomonitor/amd64: build binary for amd64
.PHONY: build/amd64
build/amd64:
	@echo 'Building cmd/vd for amd64...'
	GOOS=linux GOARCH=amd64 go build -o=./bin/amd64/vd .



