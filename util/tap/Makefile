TESTS = known auto

.PHONY: $(TESTS)

all: $(TESTS)

$(TESTS): %: test/%/main.go
	go build -o $@ test/$@/main.go
	prove -v -e '' ./$@
