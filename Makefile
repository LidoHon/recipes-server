# Include variables from the .env file
include .env
# variable declarations
current_time = $(shell date --iso-8601=seconds)
git_description = $(shell git describe --always --dirty --tags --long)
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_description}'

# =============================================================================== #
# HELPERS
# =============================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && \
	if [ -z "$$ans" ]; then \
		echo "You haven't entered anything. Please try again."; \
		false; \
	elif [ "$${ans:-N}" = "y" ]; then \
		echo "Proceeding..."; \
	else \
		echo "No? Thanks, bye!"; \
		false; \
	fi

# ============================================================================== #
# DEVELOPMENT
# ============================================================================== #

## run/api: run the cmd/api application
.PHONY: run-go-app
run-go-app:
	@cd go-app && go run .



# ============================================================================== #
# Build
# ============================================================================== # 
## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags=${linker_flags} -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/api ./cmd/api



# ============================================================================== #
# QUALITY CONTROL
# ============================================================================== # 

## audit: tidy dependancies and format,vet and test all codes

.PHONY: audit
audit: vendor
	@echo 'Tidying and verfying module dependancies...'
	cd go-app && go mod tidy
	cd go-app && go mod verify
	cd go-app && go mod vendor

	@echo 'Formatting code...'
	cd go-app && go fmt ./...

	@echo 'Vetting code...'
	cd go-app && go vet ./...
	scd go-app && staticcheck ./...

	@echo 'Running tests...'
	cd go-app && go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verfying module dependancies...'
	cd go-app && go mod tidy
	cd go-app && go mod verify

	@echo 'Vendering dependancies'
	cd go-app &&  go mod vendor