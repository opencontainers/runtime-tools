// +build linux solaris

package validate

import (
	"fmt"
	"os"
	"syscall"

	"github.com/hashicorp/go-multierror"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/specerror"
)

func (v *Validator) checkPosixFilesystemDevice(device *rspec.LinuxDevice, info os.FileInfo) (err error) {
	fStat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return multierror.Append(err, specerror.NewError(specerror.DevicesAvailable,
			fmt.Errorf("cannot determine state for device %s", device.Path), rspec.Version))
	}
	var devType string
	switch fStat.Mode & syscall.S_IFMT {
	case syscall.S_IFCHR:
		devType = "c"
	case syscall.S_IFBLK:
		devType = "b"
	case syscall.S_IFIFO:
		devType = "p"
	default:
		devType = "unmatched"
	}
	if devType != device.Type || (devType == "c" && device.Type == "u") {
		err = multierror.Append(err, specerror.NewError(specerror.DevicesFileNotMatch,
			fmt.Errorf("unmatched %s already exists in filesystem", device.Path), rspec.Version))
	}
	if devType != "p" {
		dev := fStat.Rdev
		major := (dev >> 8) & 0xfff
		minor := (dev & 0xff) | ((dev >> 12) & 0xfff00)
		if int64(major) != device.Major || int64(minor) != device.Minor {
			err = multierror.Append(err, specerror.NewError(specerror.DevicesFileNotMatch,
				fmt.Errorf("unmatched %s already exists in filesystem", device.Path), rspec.Version))
		}
	}
	if device.FileMode != nil {
		expectedPerm := *device.FileMode & os.ModePerm
		actualPerm := info.Mode() & os.ModePerm
		if expectedPerm != actualPerm {
			err = multierror.Append(err, specerror.NewError(specerror.DevicesFileNotMatch,
				fmt.Errorf("unmatched %s already exists in filesystem", device.Path), rspec.Version))
		}
	}
	if device.UID != nil {
		if *device.UID != fStat.Uid {
			err = multierror.Append(err, specerror.NewError(specerror.DevicesFileNotMatch,
				fmt.Errorf("unmatched %s already exists in filesystem", device.Path), rspec.Version))
		}
	}
	if device.GID != nil {
		if *device.GID != fStat.Gid {
			err = multierror.Append(err, specerror.NewError(specerror.DevicesFileNotMatch,
				fmt.Errorf("unmatched %s already exists in filesystem", device.Path), rspec.Version))
		}
	}

	return err
}
