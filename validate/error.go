package validate

import (
	"errors"
	"fmt"
	"strings"
)

// ComplianceLevel represents the OCI compliance levels
type ComplianceLevel int

const (
	ComplianceOptional ComplianceLevel = iota
	ComplianceMay
	ComplianceRecommended
	ComplianceShould
	ComplianceShouldNot
	ComplianceShall
	ComplianceShallNot
	ComplianceRequired
	ComplianceMustNot
	ComplianceMust
)

// OCIErrorCode represents the compliance content
type OCIErrorCode int

const (
	DefaultFilesystems OCIErrorCode = iota
)

// OCIError represents an error with compliance level and OCI reference
type OCIError struct {
	Level     ComplianceLevel
	Reference string
	Err       error
}

//FIXME: change to tagged spec releases
const referencePrefix = "https://github.com/opencontainers/runtime-spec/blob/master/"

var ociErrors = map[OCIErrorCode]OCIError{
	DefaultFilesystems: OCIError{Level: ComplianceShould, Reference: "config-linux.md#default-filesystems"},
}

// ParseLevel takes a string level and returns the OCI compliance level constant
func ParseLevel(level string) ComplianceLevel {
	switch strings.ToUpper(level) {
	case "SHOULD":
		return ComplianceShould
	case "MUST":
		return ComplianceMust
	default:
		return ComplianceMust
	}
}

// NewOCIError creates an OCIError by OCIErrorCode and message
func NewOCIError(code OCIErrorCode, msg string) error {
	err := ociErrors[code]
	err.Err = errors.New(msg)

	return &err
}

// Error returns the error message with OCI reference
func (oci *OCIError) Error() string {
	return fmt.Sprintf("%s\nRefer to: %s%s", oci.Err.Error(), referencePrefix, oci.Reference)
}
