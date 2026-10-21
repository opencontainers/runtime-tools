package util

import (
	"testing"

	"golang.org/x/sys/unix"
)

func TestLargeRlimitsSupportedRestores(t *testing.T) {
	var before, after unix.Rlimit

	if err := unix.Prlimit(0, unix.RLIMIT_FSIZE, nil, &before); err != nil {
		t.Fatal(err)
	}
	if _, err := LargeRlimitsSupported(); err != nil {
		t.Fatal(err)
	}
	if err := unix.Prlimit(0, unix.RLIMIT_FSIZE, nil, &after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Errorf("RLIMIT_FSIZE changed from %+v to %+v", before, after)
	}
}
