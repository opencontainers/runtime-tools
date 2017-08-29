package error

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

const referenceTemplate = "https://github.com/opencontainers/runtime-spec/blob/v%s/%s"

// SpecErrorCode represents the compliance content.
type SpecErrorCode int

const (
	// NonError represents that an input is not an error
	NonError SpecErrorCode = iota
	// NonRFCError represents that an error is not a rfc2119 error
	NonRFCError

	// ConfigFileExistence represents the error code of 'config.json' existence test
	ConfigFileExistence
	// ArtifactsInSingleDir represents the error code of artifacts place test
	ArtifactsInSingleDir

	// SpecVersion represents the error code of specfication version test
	SpecVersion

	// RootOnNonHyperV represents the error code of root setting test on non hyper-v containers
	RootOnNonHyperV
	// RootOnHyperV represents the error code of root setting test on hyper-v containers
	RootOnHyperV
	// PathFormatOnWindows represents the error code of the path format test on Window
	PathFormatOnWindows
	// PathName represents the error code of the path name test
	PathName
	// PathExistence represents the error code of the path existence test
	PathExistence
	// ReadonlyFilesystem represents the error code of readonly test
	ReadonlyFilesystem
	// ReadonlyOnWindows represents the error code of readonly setting test on Windows
	ReadonlyOnWindows

	// DefaultFilesystems represents the error code of default filesystems test
	DefaultFilesystems
)

type errorTemplate struct {
	Level     Level
	Reference func(version string) (reference string, err error)
}

var (
	containerFormatRef = func(version string) (reference string, err error) {
		return fmt.Sprintf(referenceTemplate, version, "bundle.md#container-format"), nil
	}
	specVersionRef = func(version string) (reference string, err error) {
		return fmt.Sprintf(referenceTemplate, version, "config.md#specification-version"), nil
	}
	rootRef = func(version string) (reference string, err error) {
		return fmt.Sprintf(referenceTemplate, version, "config.md#root"), nil
	}
	defaultFSRef = func(version string) (reference string, err error) {
		return fmt.Sprintf(referenceTemplate, version, "config-linux.md#default-filesystems"), nil
	}
)

var ociErrors = map[SpecErrorCode]errorTemplate{
	// Bundle.md
	// Container Format
	ConfigFileExistence:  errorTemplate{Level: Must, Reference: containerFormatRef},
	ArtifactsInSingleDir: errorTemplate{Level: Must, Reference: containerFormatRef},

	// Config.md
	// Specification Version
	SpecVersion: errorTemplate{Level: Must, Reference: specVersionRef},
	// Root
	RootOnNonHyperV: errorTemplate{Level: Required, Reference: rootRef},
	RootOnHyperV:    errorTemplate{Level: Must, Reference: rootRef},
	// TODO: add tests for 'PathFormatOnWindows'
	PathFormatOnWindows: errorTemplate{Level: Must, Reference: rootRef},
	PathName:            errorTemplate{Level: Should, Reference: rootRef},
	PathExistence:       errorTemplate{Level: Must, Reference: rootRef},
	ReadonlyFilesystem:  errorTemplate{Level: Must, Reference: rootRef},
	ReadonlyOnWindows:   errorTemplate{Level: Must, Reference: rootRef},

	// Config-Linux.md
	// Default Filesystems
	DefaultFilesystems: errorTemplate{Level: Should, Reference: defaultFSRef},
}

// NewError creates an Error referencing a spec violation.  The error
// can be cast to a *runtime-tools.error.Error for extracting
// structured information about the level of the violation and a
// reference to the violated spec condition.
//
// A version string (for the version of the spec that was violated)
// must be set to get a working URL.
func NewError(code SpecErrorCode, msg string, version string) (err error) {
	template := ociErrors[code]
	reference, err := template.Reference(version)
	if err != nil {
		return err
	}
	return &Error{
		Level:     template.Level,
		Reference: reference,
		Err:       errors.New(msg),
		ErrCode:   int(code),
	}
}

// FindError finds an error from a source error (mulitple error) and
// returns the error code if founded.
// If the source error is nil or empty, return NonError.
// If the source error is not a multiple error, return NonRFCError.
func FindError(err error, code SpecErrorCode) SpecErrorCode {
	if err == nil {
		return NonError
	}

	if merr, ok := err.(*multierror.Error); ok {
		if merr.ErrorOrNil() == nil {
			return NonError
		}
		for _, e := range merr.Errors {
			if rfcErr, ok := e.(*Error); ok {
				if rfcErr.ErrCode == int(code) {
					return code
				}
			}
		}
	}
	return NonRFCError
}
