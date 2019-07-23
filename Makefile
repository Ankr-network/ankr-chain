BUILD_TAGS?='ankrchain'
OUTPUT?=build/ankrchain
BUILD_FLAGS = -ldflags "-X github.com/tendermint/tendermint/version.GitCommit=`git rev-parse --short=8 HEAD`"

OUTPUTTOOLDIR?=build/tool

export GO111MODULE=on

all: build install

build:
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -tags $(BUILD_TAGS) -o $(OUTPUT) ./main.go

install:
	CGO_ENABLED=0 go install  $(BUILD_FLAGS) -tags $(BUILD_TAGS) ./main.go

tools:
	CGO_ENABLED=0 go build  -o $OUTPUTTOOLDIR/base64show  ./tool/base64_show.go
	CGO_ENABLED=0 go build  -o $OUTPUTTOOLDIR/keygen      ./tool/keygen.go

fmt:
	@go fmt ./...

lint:
	@echo "--> Running linter"
	@golangci-lint run

.PHONY: check build install fmt lint

