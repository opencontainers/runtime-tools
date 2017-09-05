package specerror

import (
	"fmt"

	rfc2119 "github.com/opencontainers/runtime-tools/error"
)

// define error codes
const (
	// DefaultFilesystems represents "The following filesystems SHOULD be made available in each container's filesystem:"
	DefaultFilesystems = "The following filesystems SHOULD be made available in each container's filesystem:"
)

var (
	defaultFilesystemsRef = func(version string) (reference string, err error) {
		return fmt.Sprintf(referenceTemplate, version, "config-linux.md#default-filesystems"), nil
	}
)

func init() {
	register(DefaultFilesystems, rfc2119.Should, defaultFilesystemsRef)
}
