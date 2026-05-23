//go:build windows

package fs

func fileOwnerSys(FileInfo) (uid, gid string, ok bool) {
	return "0", "0", false
}
