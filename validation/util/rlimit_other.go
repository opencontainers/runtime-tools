//go:build !linux

package util

// LargeRlimitsSupported reports whether the host kernel holds resource
// limits of 4 GiB and more.
func LargeRlimitsSupported() (bool, error) {
	return true, nil
}
