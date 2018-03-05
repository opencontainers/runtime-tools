package validate

import (
	"runtime"

	"github.com/sirupsen/logrus"
)

func (v *Validator) checkPosixFilesystemDevice(device *rspec.LinuxDevice, info os.FileInfo) (err error) {
	logrus.Warnf("checking POSIX devices is not supported on %s", runtime.GOOS)
	return nil
}
