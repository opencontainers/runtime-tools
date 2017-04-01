package validate

import (
	"errors"
	"fmt"
	"strings"
)

// ComplianceLevel represents the OCI compliance levels
type ComplianceLevel int

const (
	// MAY-level
	ComplianceMay ComplianceLevel = iota
	ComplianceOptional
	// SHOULD-level
	ComplianceShould
	ComplianceShouldNot
	ComplianceRecommended
	ComplianceNotRecommended
	// MUST-level
	ComplianceMust
	ComplianceMustNot
	ComplianceShall
	ComplianceShallNot
	ComplianceRequired
)

// ErrorCode represents the compliance content
type ErrorCode int

const (
	DefaultFilesystems ErrorCode = iota
)

// Error represents an error with compliance level and OCI reference
type Error struct {
	Level     ComplianceLevel
	Reference string
	Err       error
}

//FIXME: change to tagged spec releases
const referencePrefix = "https://github.com/opencontainers/runtime-spec/blob/master/"

var ociErrors = map[ErrorCode]Error{
	DefaultFilesystems: Error{Level: ComplianceShould, Reference: "config-linux.md#default-filesystems"},
}

// ParseLevel takes a string level and returns the OCI compliance level constant
func ParseLevel(level string) ComplianceLevel {
	switch strings.ToUpper(level) {
	case "MAY":
		fallthrough
	case "OPTIONAL":
		return ComplianceMay
	case "SHOULD":
		fallthrough
	case "SHOULDNOT":
		fallthrough
	case "RECOMMENDED":
		fallthrough
	case "NOTRECOMMENDED":
		return ComplianceShould
	case "MUST":
		fallthrough
	case "MUSTNOT":
		fallthrough
	case "SHALL":
		fallthrough
	case "SHALLNOT":
		fallthrough
	case "REQUIRED":
		return ComplianceMust
	default:
		return ComplianceMust
	}
}

// NewError creates an Error by ErrorCode and message
func NewError(code ErrorCode, msg string) error {
	err := ociErrors[code]
	err.Err = errors.New(msg)

	return &err
}

// Error returns the error message with OCI reference
func (oci *Error) Error() string {
	return fmt.Sprintf("%s\nRefer to: %s%s", oci.Err.Error(), referencePrefix, oci.Reference)
}
