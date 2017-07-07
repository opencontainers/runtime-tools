PREFIX ?= $(DESTDIR)/usr
BINDIR ?= $(DESTDIR)/usr/bin

BUILDTAGS=
RUNTIME ?= runc
COMMIT=$(shell git rev-parse HEAD 2> /dev/null || true)

all: tool runtimetest

tool:
	go build -tags "$(BUILDTAGS)" -ldflags "-X main.gitCommit=${COMMIT}" -o oci-runtime-tool ./cmd/oci-runtime-tool

.PHONY: runtimetest
runtimetest:
	go build -tags "$(BUILDTAGS)" -o runtimetest ./cmd/runtimetest

.PHONY: man
man:
	go-md2man -in "man/oci-runtime-tool.1.md" -out "oci-runtime-tool.1"
	go-md2man -in "man/oci-runtime-tool-generate.1.md" -out "oci-runtime-tool-generate.1"
	go-md2man -in "man/oci-runtime-tool-validate.1.md" -out "oci-runtime-tool-validate.1"

install: man
	install -d -m 755 $(BINDIR)
	install -m 755 oci-runtime-tool $(BINDIR)
	install -d -m 755 $(PREFIX)/share/man/man1
	install -m 644 *.1 $(PREFIX)/share/man/man1
	install -d -m 755 $(PREFIX)/share/bash-completion/completions
	install -m 644 completions/bash/oci-runtime-tool $(PREFIX)/share/bash-completion/completions

uninstall:
	rm -f $(BINDIR)/oci-runtime-tool
	rm -f $(PREFIX)/share/man/man1/oci-runtime-tool*.1
	rm -f $(PREFIX)/share/bash-completion/completions/oci-runtime-tool

clean:
	rm -f oci-runtime-tool runtimetest *.1

localvalidation: runtimetest
	RUNTIME=$(RUNTIME) go test -tags "$(BUILDTAGS)" ${TESTFLAGS} -v github.com/opencontainers/runtime-tools/validation

.PHONY: test .gofmt .govet .golint

PACKAGES = $(shell go list ./... | grep -v vendor)
test: .gofmt .govet .golint .gotest

.gofmt:
	OUT=$$(go fmt $(PACKAGES)); if test -n "$${OUT}"; then echo "$${OUT}" && exit 1; fi

.govet:
	go vet -x $(PACKAGES)

.golint:
	golint -set_exit_status $(PACKAGES)

UTDIRS = ./validate/...
.gotest:
	go test $(UTDIRS)
