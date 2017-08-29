package validate

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	rfc2119 "github.com/opencontainers/runtime-tools/error"
)

const referenceTemplate = "https://github.com/opencontainers/runtime-spec/blob/v%s/%s"

// ErrorCode represents the compliance content.
type ErrorCode int

const (
	// NonError represents that an input is not an error
	NonError ErrorCode = iota
	// NonRFCError represents that an error is not a rfc2119 error
	NonRFCError

	// ConfigFileExistence represents the error code of 'config.json' existence test.
	ConfigFileExistence
	// ArtifactsInSignleDir represents the error code of artifacts place test.
	ArtifactsInSingleDir

	// SpecVersion represents the error code of specfication version test.
	SpecVersion

	// RootOnNonHyperV represents the error code of root setting test on non hyper-v containers
	RootOnNonHyperV
	// RootOnNonHyperV represents the error code of root setting test on hyper-v containers
	RootOnHyperV
	// PathFormatOnwindows represents the error code of the path format test on Window
	PathFormatOnWindows
	// PathName represents the error code of the path name test
	PathName
	// PathExistence represents the error code of the path existence test
	PathExistence
	//  ReadonlyFilesystem represents the error code of readonly test
	ReadonlyFilesystem
	// ReadonlyOnWindows represents the error code of readonly setting test on Windows
	ReadonlyOnWindows

	// DefaultFilesystems represents the error code of default filesystems test.
	DefaultFilesystems
)

type errorTemplate struct {
	Level     rfc2119.Level
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

var ociErrors = map[ErrorCode]errorTemplate{
	// NonRFCError represents that an error is not a rfc2119 error
	// Bundle.md
	// Container Format
	ConfigFileExistence:  errorTemplate{Level: rfc2119.Must, Reference: containerFormatRef},
	ArtifactsInSingleDir: errorTemplate{Level: rfc2119.Must, Reference: containerFormatRef},

	// Config.md
	// Specification Version
	SpecVersion: errorTemplate{Level: rfc2119.Must, Reference: specVersionRef},
	// Root
	RootOnNonHyperV: errorTemplate{Level: rfc2119.Required, Reference: rootRef},
	RootOnHyperV:    errorTemplate{Level: rfc2119.Must, Reference: rootRef},
	// TODO
	PathFormatOnWindows: errorTemplate{Level: rfc2119.Must, Reference: rootRef},
	PathName:            errorTemplate{Level: rfc2119.Should, Reference: rootRef},
	PathExistence:       errorTemplate{Level: rfc2119.Must, Reference: rootRef},
	ReadonlyFilesystem:  errorTemplate{Level: rfc2119.Must, Reference: rootRef},
	ReadonlyOnWindows:   errorTemplate{Level: rfc2119.Must, Reference: rootRef},

	// Config-Linux.md
	// Default Filesystems
	DefaultFilesystems: errorTemplate{Level: rfc2119.Should, Reference: defaultFSRef},
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
		ErrCode:   int(code),
	}
}

// FindError finds an error from a source error (mulitple error) and
// returns the error code if founded.
// If the source error is nil or empty, return NonErr.
// If the source error is not a multiple error, return NonRFCErr.
func FindError(err error, code ErrorCode) ErrorCode {
	if err == nil {
		return NonError
	}

	if merr, ok := err.(*multierror.Error); ok {
		if merr.ErrorOrNil() == nil {
			return NonError
		}
		for _, e := range merr.Errors {
			if rfcErr, ok := e.(*rfc2119.Error); ok {
				if rfcErr.ErrCode == int(code) {
					return code
				}
			}
		}
	}
	return NonRFCError
}
