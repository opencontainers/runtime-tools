TESTS = known auto

.PHONY: $(TESTS)

all: $(TESTS)

$(TESTS): %: test/%/main.go
	go build -o test/$@/test test/$@/main.go
	prove -v -e '' test/$@/test
