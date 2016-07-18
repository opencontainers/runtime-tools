PREFIX ?= $(DESTDIR)/usr
BINDIR ?= $(DESTDIR)/usr/bin

BUILDTAGS=
export GOPATH:=$(CURDIR)/Godeps/_workspace:$(GOPATH)

all:
	go build -tags "$(BUILDTAGS)" -o ocitools ./cmd/ocitools
	go build -tags "$(BUILDTAGS)" -o runtimetest ./cmd/runtimetest

.PHONY: man
man:
	go-md2man -in "man/ocitools.1.md" -out "ocitools.1"
	go-md2man -in "man/ocitools-generate.1.md" -out "ocitools-generate.1"
	go-md2man -in "man/ocitools-validate.1.md" -out "ocitools-validate.1"

install: man
	install -d -m 755 $(BINDIR)
	install -m 755 ocitools $(BINDIR)
	install -d -m 755 $(PREFIX)/share/man/man1
	install -m 644 *.1 $(PREFIX)/share/man/man1
	install -d -m 755 $(PREFIX)/share/bash-completion/completions
	install -m 644 completions/bash/ocitools $(PREFIX)/share/bash-completion/completions

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

