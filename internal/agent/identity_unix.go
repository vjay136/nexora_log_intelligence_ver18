//go:build !windows

package agent

import (
	"os"
	"syscall"
)

func identity(_ *os.File, info os.FileInfo) (uint64, uint64, error) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, uint64(info.ModTime().UnixNano()), nil
	}
	return uint64(stat.Dev), stat.Ino, nil
}
