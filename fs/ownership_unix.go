//go:build !windows

package fs

import (
	"strconv"
	"syscall"
)

func fileOwnerSys(stat FileInfo) (uid, gid string, ok bool) {
	if sys, ok := stat.Sys().(*syscall.Stat_t); ok {
		return strconv.Itoa(int(sys.Uid)), strconv.Itoa(int(sys.Gid)), true
	}

	return "0", "0", false
}
