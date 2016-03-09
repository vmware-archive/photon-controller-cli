# Makefile to help building go components

GO=godep go
COMMAND_NAME=photon
LDFLAGS="-X main.githash=`git rev-parse --short HEAD` -X main.commandName=$(COMMAND_NAME)"
GOBUILD=$(GO) build -ldflags $(LDFLAGS)

all: test build binaries

binaries: darwin/amd64 windows/amd64 linux/amd64

darwin/amd64:
	$(eval export GOOS=darwin)
	$(eval export GOARCH=amd64)
	$(eval export fileext=)
	make build

windows/amd64:
	$(eval export GOOS=windows)
	$(eval export GOARCH=amd64)
	$(eval export fileext=.exe)
	make build

linux/amd64:
	$(eval export GOOS=linux)
	$(eval export GOARCH=amd64)
	$(eval export fileext=)
	make build

# go build arch is controlled by env var GOOS and GOARCH, when not set it use current machine native arch
build:
	$(GOBUILD) -o bin/$(GOOS)$(GOARCH)/$(COMMAND_NAME)$(fileext) ./photon

#
# get the tools
#
tools:
	go get -u github.com/kisielk/errcheck
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/golang/lint/golint
	go get -u github.com/tools/godep

test: tools
	errcheck ./...
	$(GO) vet ./...
	golint
	! gofmt -l photon 2>&1 | read || (gofmt -d photon; echo "ERROR: Fix gofmt errors. Run 'gofmt -w photon'"; exit 1)
	$(GO) test -v ./...
