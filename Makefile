
include .envrc

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## db/migration/new: create migration files with specified filename
.PHONY: db/migration/new
db/migration/new:
	@read  -p "Input file name: " filename; \
	migrate create -seq -ext .sql -dir ./migrations $$filename

## db/migration/apply: apply the migration with the [ up | down | goto # | force # ] as specified
.PHONY: db/migration/apply
db/migration/apply:
	@read -p "Input apply params: " apply_params; \
	migrate -path ./migrations -database ${LETSCHAT_API_DB_DSN} $$apply_params

## compose/run: run docker compose with your specified command & flags
.PHONY: compose/run
compose/run:
	@read -p "Input command with flags: " command; \
	docker compose -f compose-dev.yml $$command

## build/debug: build with specific flags that allows delve debugging on remote port
.PHONY: build/debug
build/debug:
	CGOENABLED=1; \
	go build -gcflags "all=-N -l" -o ./bin ./cmd/letschat; \
	dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./bin/letschat.exe -- -usr 2

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Formating code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	CGO_ENABLED=1 go test -race -vet=off ./...

# ==================================================================================== #
# BUILD
# ==================================================================================== #

current_time = $(shell date --iso-8601=seconds)
git_description = $(shell git describe --always --dirty --tags --long)
linker_flags = '-s -w -X main.buildTime=${current_time} -X main.version=${git_description}'

## build/letschat: build the letschat TUI binary with compression using LZMA
.PHONY: build/letschat
build/letschat:
	mkdir -p bin && \
 	go build -ldflags="-s -w" -trimpath -o bin/letschat.exe ./cmd/letschat && \
 	upx --best --lzma bin/letschat.exe

## build/letschat-api: build the letschat API binary with compression using LZMA
.PHONY: build/letschat-api
build/letschat-api:
	GOOS=windows GOARCH=amd64 go build -ldflags=${linker_flags} -o ./bin/letschat-api.exe ./cmd/letschat-api && \
	upx --best --lzma bin/letschat-api.exe
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o ./bin/letschat-api ./cmd/letschat-api

## build/docker: build the letschat API binary for linux
.PHONY: build/docker
build/docker:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o ./bin/letschat-api ./cmd/letschat-api