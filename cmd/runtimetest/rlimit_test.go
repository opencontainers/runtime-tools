package main

import (
	"syscall"
	"testing"

	"golang.org/x/sys/unix"
)

func TestParseNofileLimits(t *testing.T) {
	limits := `Limit                     Soft Limit           Hard Limit           Units
Max processes             63432                63432                processes
Max open files            3000                 4000                 files
Max locked memory         8388608              8388608              bytes
`
	rlimit, err := parseNofileLimits(limits)
	if err != nil {
		t.Fatal(err)
	}
	if rlimit.Cur != 3000 || rlimit.Max != 4000 {
		t.Errorf("limits == %d/%d, expected 3000/4000", rlimit.Cur, rlimit.Max)
	}

	rlimit, err = parseNofileLimits("Max open files            unlimited            unlimited            files\n")
	if err != nil {
		t.Fatal(err)
	}
	if rlimit.Cur != unix.RLIM_INFINITY || rlimit.Max != unix.RLIM_INFINITY {
		t.Errorf("limits == %d/%d, expected unlimited", rlimit.Cur, rlimit.Max)
	}

	if _, err := parseNofileLimits("Max processes 1 1 processes\n"); err == nil {
		t.Error("expected an error without a RLIMIT_NOFILE entry")
	}
}

func TestGetRlimitNofile(t *testing.T) {
	var own syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &own); err != nil {
		t.Fatal(err)
	}

	rlimit, err := getRlimit(syscall.RLIMIT_NOFILE)
	if err != nil {
		t.Skipf("cannot read the limits of a child: %v", err)
	}
	// Go only raises the soft limit, never the hard one.
	if rlimit.Max != own.Max {
		t.Errorf("hard limit == %d, expected %d", rlimit.Max, own.Max)
	}
	if rlimit.Cur > own.Cur {
		t.Errorf("soft limit %d of a child exceeds our own %d", rlimit.Cur, own.Cur)
	}
}
