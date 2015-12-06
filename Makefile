BUILDTAGS=
export GOPATH:=$(CURDIR)/Godeps/_workspace:$(GOPATH)

all:
	go build -tags "$(BUILDTAGS)" -o ocitools .
	go build -tags "$(BUILDTAGS)" -o runtimetest ./cmd/runtimetest

install:
	cp ocitools /usr/local/bin/ocitools

rootfs/bin/busybox: downloads/stage3-amd64-current.tar.bz2 rootfs-files
	gpg --verify $<.DIGESTS.asc
	(cd downloads && \
		grep -A1 '^# SHA512 HASH' stage3-amd64-current.tar.bz2.DIGESTS.asc | \
		grep -v '^--' | \
		sha512sum -c)
	sudo rm -rf rootfs
	sudo mkdir rootfs
	sudo tar -xvf downloads/stage3-amd64-current.tar.bz2 -C rootfs \
		--no-recursion --wildcards $$(< rootfs-files)
	sudo touch $@

downloads/stage3-amd64-current.tar.bz2: get-stage3.sh
	./$<
	touch downloads/stage3-amd64-*.tar.bz2

clean:
	rm -f ocitools runtimetest downloads/*
	sudo rm -rf rootfs
