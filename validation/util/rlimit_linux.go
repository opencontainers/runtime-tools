package util

import (
	"errors"

	"golang.org/x/sys/unix"
)

// LargeRlimitsSupported reports whether resource limits of 4 GiB and more
// can be set and read back on this host. A 32-bit kernel stores them in an
// unsigned long and turns any larger value into RLIM_INFINITY, and without
// CAP_SYS_RESOURCE a hard limit below 4 GiB cannot be raised to them.
func LargeRlimitsSupported() (bool, error) {
	const probe = 1 << 32
	var old, cur unix.Rlimit

	if err := unix.Prlimit(0, unix.RLIMIT_FSIZE, nil, &old); err != nil {
		return false, err
	}

	cur = unix.Rlimit{Cur: probe, Max: max(old.Max, probe)}
	if err := unix.Prlimit(0, unix.RLIMIT_FSIZE, &cur, nil); err != nil {
		if errors.Is(err, unix.EPERM) {
			return false, nil
		}
		return false, err
	}
	defer unix.Prlimit(0, unix.RLIMIT_FSIZE, &old, nil)

	if err := unix.Prlimit(0, unix.RLIMIT_FSIZE, nil, &cur); err != nil {
		return false, err
	}
	return cur.Cur == probe, nil
}
