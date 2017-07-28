package validate

import (
	"errors"
	"fmt"

	rfc2119 "github.com/opencontainers/runtime-tools/error"
)

const referenceTemplate = "https://github.com/opencontainers/runtime-spec/blob/v%s/%s"

// ErrorCode represents the compliance content.
type ErrorCode int

const (
	// DefaultFilesystems represents the error code of default filesystems test.
	DefaultFilesystems ErrorCode = iota
)

type errorTemplate struct {
	Level     rfc2119.Level
	Reference func(version string) (reference string, err error)
}

var ociErrors = map[ErrorCode]errorTemplate{
	DefaultFilesystems: errorTemplate{
		Level: rfc2119.Should,
		Reference: func(version string) (reference string, err error) {
			return fmt.Sprintf(referenceTemplate, version, "config-linux.md#default-filesystems"), nil
		},
	},
}

// NewError creates an Error referencing a spec violation.  The error
// can be cast to a *runtime-tools.error.Error for extracting
// structured information about the level of the violation and a
// reference to the violated spec condition.
//
// A version string (for the version of the spec that was violated)
// must be set to get a working URL.
func NewError(code ErrorCode, msg string, version string) (err error) {
	template := ociErrors[code]
	reference, err := template.Reference(version)
	if err != nil {
		return err
	}
	return &rfc2119.Error{
		Level:     template.Level,
		Reference: reference,
		Err:       errors.New(msg),
	}
}
