// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

package fs

type SymlinkFS interface {
	FS

	// Readlink returns the destination of the named symbolic link. If there is
	// an error, it will be of type *PathError.
	//
	// If link is out of filesystem bounds (e.g. link at `a/b.link` forwards to
	// `../../../../../sensitive.data`, then this method should throw PathError)
	Readlink(name string) (string, error)

	// Lstat returns a FileInfo describing the named file. If the file is a
	// symbolic link, the returned FileInfo describes the symbolic link. Lstat
	// makes no attempt to follow the link. If there is an error, it will be of
	// type *PathError.
	Lstat(name string) (FileInfo, error)
}

type SymlinkWFS interface {
	WFS
	SymlinkFS

	// Symlink creates newname as a symbolic link to oldname. If there is an
	// error, it will be of type *LinkError.
	Symlink(oldname, newname string) error
}
