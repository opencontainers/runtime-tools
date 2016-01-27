BUILDTAGS=
export GOPATH:=$(CURDIR)/Godeps/_workspace:$(GOPATH)

all:
	go build -tags "$(BUILDTAGS)" -o ocitools .
	go build -tags "$(BUILDTAGS)" -o runtimetest ./cmd/runtimetest

install:
	cp ocitools /usr/local/bin/ocitools

clean:
	rm -f ocitools runtimetest

.PHONY: test .gofmt .govet .golint

test: .gofmt .govet .golint

.gofmt:
	go fmt ./...

.govet:
	go vet -x ./...

.golint:
	golint ./...

