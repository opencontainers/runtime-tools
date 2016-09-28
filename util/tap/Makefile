TESTS = known auto

.PHONY: $(TESTS)

all: $(TESTS)

clean:
	rm -f test/*/test

$(TESTS): %: test/%/main.go
	go build -o test/$@/test test/$@/main.go
	prove -v -e '' test/$@/test
