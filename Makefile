BUILDTAGS=
export GOPATH:=$(CURDIR)/Godeps/_workspace:$(GOPATH)

all:
	go build -tags "$(BUILDTAGS)" -o ocitools .
	go build -tags "$(BUILDTAGS)" -o runtimetest ./cmd/runtimetest

install:
	cp ocitools /usr/local/bin/ocitools

downloads/stage3-amd64-current.tar.bz2: get-stage3.sh
	./$<
	touch downloads/stage3-amd64-*.tar.bz2

clean:
	rm -f ocitools runtimetest downloads/*
