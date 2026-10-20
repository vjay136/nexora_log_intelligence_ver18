//go:build windows

package agent

import (
	"os"
	"syscall"
)

// Windows exposes a stable volume serial number and file index for an open
// handle. Modification time cannot be used as identity: appending to a file
// changes it and would make a normal append look like rotation, replaying the
// file from offset zero on every restart.
func identity(file *os.File, _ os.FileInfo) (uint64, uint64, error) {
	var info syscall.ByHandleFileInformation
	if err := syscall.GetFileInformationByHandle(syscall.Handle(file.Fd()), &info); err != nil {
		return 0, 0, err
	}
	fileIndex := uint64(info.FileIndexHigh)<<32 | uint64(info.FileIndexLow)
	return uint64(info.VolumeSerialNumber), fileIndex, nil
}
