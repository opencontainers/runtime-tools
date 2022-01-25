PREFIX ?= $(DESTDIR)/usr
BINDIR ?= $(DESTDIR)/usr/bin
TAPTOOL ?= tap

BUILDTAGS=
RUNTIME ?= runc
COMMIT ?= $(shell git describe --dirty --long --always --tags 2> /dev/null)
VERSION := ${shell cat ./VERSION}
BUILD_FLAGS := -tags "$(BUILDTAGS)" -ldflags "-X main.gitCommit=$(COMMIT) -X main.version=$(VERSION)" $(EXTRA_FLAGS)
STATIC_BUILD_FLAGS := -tags "$(BUILDTAGS) netgo osusergo" -ldflags "-extldflags -static -X main.gitCommit=$(COMMIT) -X main.version=$(VERSION)" $(EXTRA_FLAGS)
VALIDATION_TESTS ?= $(patsubst %.go,%.t,$(shell find ./validation/ -name *.go | grep -v util))

all: tool runtimetest validation-executables

tool:
	go build $(BUILD_FLAGS) -o oci-runtime-tool ./cmd/oci-runtime-tool

.PHONY: runtimetest
runtimetest:
	go build $(STATIC_BUILD_FLAGS) -o runtimetest ./cmd/runtimetest

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
	rm -f oci-runtime-tool runtimetest *.1 $(VALIDATION_TESTS)

localvalidation:
	@for EXECUTABLE in runtimetest $(VALIDATION_TESTS); \
	do \
		if test ! -x "$${EXECUTABLE}"; \
		then \
			echo "missing test executable $${EXECUTABLE}; run 'make runtimetest validation-executables'" >&2; \
			exit 1; \
		fi; \
	done
	RUNTIME=$(RUNTIME) $(TAPTOOL) $(VALIDATION_TESTS)

.PHONY: validation-executables
validation-executables: $(VALIDATION_TESTS)

.PRECIOUS: $(VALIDATION_TESTS)
.PHONY: $(VALIDATION_TESTS)
$(VALIDATION_TESTS): %.t: %.go
	go build $(BUILD_FLAGS) -o $@ $<

print-validation-tests:
	@echo $(VALIDATION_TESTS)

.PHONY: test .govet print-validation-tests

PACKAGES = $(shell go list ./... | grep -v vendor)
test: .govet .gotest

.govet:
	go vet -x $(PACKAGES)

.gotest:
	go test $(TESTFLAGS) ./...
