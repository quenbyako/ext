// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

// wfs file contains WFS abstraction, which allows you not only to
// read from io/fs, but also to write them. This is important to
// make your code abstracted from real os functionality and
// create stable tests. You can use WFS for example to interact
// with KV databases, use ETCD as
// filesystem, and more useful and fun stuff.
//
// But! Main purpose of this implementation is to detach tests
// from local machine, so if you want to add some implementation
// — you're welcome, but don't expect that WFS will be more
// complex than default io/fs package.
//
// References:
// - https://github.com/golang/go/issues/45757
// - https://github.com/psanford/memfs
// - https://goplay.tools/snippet/6X1A8oXoNX-
package fs

import (
	"os"
)

// WFS is an implementation of filesystem in write mode.
//
// If filesystem supports permissions for each node (read, write, execute and
// other permissions), when it creates new file, it should use default
// permissions rw-r--r--. It also can implement [WriteFileFS] interface, to
// create files with more precise permissions
//
// If filesystem supports symlinks, it should implement [SymlinkFS] interface.
type WFS interface {
	FS

	// OpenW open file with write mode. If file not exist, it should be created.
	OpenW(name string) (WFile, error)

	// Remove removes file from filesystem. If file not exists, The method must
	// throw *fs.PathError[fs.ErrNotExist]
	Remove(name string) error
}

type WriteFileFS interface {
	FS

	WriteFile(name string, data []byte, perm FileMode) error
}

func WriteFile(fsys WFS, name string, data []byte, perm FileMode) error {
	if fsys, ok := fsys.(WriteFileFS); ok {
		return fsys.WriteFile(name, data, perm)
	}

	f, err := fsys.OpenW(name)
	if err != nil {
		return err
	}
	defer func() {
		if err1 := f.Close(); err1 != nil && err == nil {
			err = err1
		}
	}()

	if _, err = f.Write(data); err != nil {
		return err
	}

	if f, ok := f.(WFileSync); ok {
		return f.Sync()
	}

	return nil
}

// WFile is an interface of File, but in write mode.
type WFile interface {
	Stat() (FileInfo, error)
	Write(p []byte) (n int, err error)
	Close() error
}

// WFileSync provides Sync method, which is usual for os implementations of
// writing files.
type WFileSync interface {
	WFile
	Sync() error
}

// to be sure that these interfaces compatible with stdlib, we are checking this
// assertion
var _ WFileSync = (*os.File)(nil)

// FileOwner is an extension for [FileInfo] to provide information about
// ownership. This is not a best practice, so use this interface in your custom
// filesystems. And very carefully!
type FileInfoOwner interface {
	FileInfo

	Owner() (uid, gid string)
}
