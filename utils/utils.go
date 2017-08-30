package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/syndtr/gocapability/capability"
)

// LastCap return last cap of system
func LastCap() capability.Cap {
	last := capability.CAP_LAST_CAP
	// hack for RHEL6 which has no /proc/sys/kernel/cap_last_cap
	if last == capability.Cap(63) {
		last = capability.CAP_BLOCK_SUSPEND
	}

	return last
}

// UnitListValid checks strings whether is valid for
// cpuset.cpus and cpuset.mems, duplicates are allowed
// Supported formats:
// 1
// 0-3
// 0-2,1,3
// 0-2,1-3,4
func UnitListValid(val string) error {
	if val == "" {
		return nil
	}

	split := strings.Split(val, ",")
	errInvalidFormat := fmt.Errorf("invalid format: %s", val)

	for _, r := range split {
		if !strings.Contains(r, "-") {
			_, err := strconv.Atoi(r)
			if err != nil {
				return errInvalidFormat
			}
		} else {
			split := strings.SplitN(r, "-", 2)
			min, err := strconv.Atoi(split[0])
			if err != nil {
				return errInvalidFormat
			}
			max, err := strconv.Atoi(split[1])
			if err != nil {
				return errInvalidFormat
			}
			if max < min {
				return errInvalidFormat
			}
		}
	}
	return nil
}
