TESTS = auto check known
GOPATH = $(CURDIR)/gopath

.PHONY: $(TESTS)

all: test/*/test
	prove -v -e '' test/*/test

clean:
	rm -f test/*/test

test/%/test: test/%/main.go
	go build -o $@ $<

$(TESTS): %: test/%/test
	prove -v -e '' test/$@/test
