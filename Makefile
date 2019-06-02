.PHONY: test lint fmt clean

OUTPUT=argo
BUILDCMD=CGO_ENABLED=0 go build -v
TESTCMD=go test -v -race

build:
	$(BUILDCMD) -ldflags '-X main.VersionSuffix=$(shell git rev-parse HEAD)' -o $(OUTPUT) cmd/argo/*.go

test:
	$(TESTCMD) ./...

lint:
	golint `go list ./... | grep -v /vendor/`

fmt:
	! go fmt ./... 2>&1 | tee /dev/tty | read

clean:
	go clean ./...
