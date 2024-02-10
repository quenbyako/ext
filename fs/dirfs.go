// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DirFS is a drop-in replacement for [os.DirFS]
//
// For symlinks, DirFS **ALWAYS** create relative paths, due to unstable and not
// standardized symlink api. If you want to create hard link, or symlink with
// global path, please, use [os] package. One way, or another, writing to fs
// interface or making symlinks is not official thing.
func DirFS(path string) SymlinkWFS {
	if path, ok := normalizeDirFS(path); ok {
		return dirFS(path)
	}

	panic(fmt.Errorf("Path %q is not absolute.", path))
}

func normalizeDirFS(path string) (_ string, isAbsolute bool) {
	if path != "" && !filepath.IsAbs(path) {
		return "", false
	} else if path == "" { // fallback if we want' to pass root dir implicitly
		path = "/"
	} else if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	return path, true
}

type dirFS string

var _ WFS = DirFS("")

func (dir dirFS) Open(name string) (File, error) { //cover:ignore
	if err := checkname(name, "open"); err != nil {
		return nil, err
	}

	f, err := os.Open(string(dir) + "/" + name)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (dir dirFS) path(name string) string { return filepath.Join(string(dir), name) }

func (dir dirFS) Stat(name string) (FileInfo, error) { //cover:ignore
	if err := checkname(name, "stat"); err != nil {
		return nil, err
	}

	f, err := os.Stat(dir.path(name))
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (dir dirFS) OpenW(name string) (WFile, error) { //cover:ignore
	if err := checkname(name, "open"); err != nil {
		return nil, err
	}

	var err error
	var f *os.File
	if _, statErr := dir.Stat(name); os.IsNotExist(statErr) {
		f, err = os.Create(name)
	} else {
		f, err = os.OpenFile(string(dir)+"/"+name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0)
	}
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (dir dirFS) WriteFile(name string, data []byte, mode FileMode) error {
	if err := checkname(name, "write"); err != nil {
		return err
	}

	dirPerms := mode
	// adding exec to only that perm groups that have any rw perm
	if mode&(ModePermUserRead|ModePermUserWrite) > 0 {
		dirPerms |= ModePermUserExec
	}
	if mode&(ModePermGroupRead|ModePermGroupWrite) > 0 {
		dirPerms |= ModePermGroupExec
	}
	if mode&(ModePermOtherRead|ModePermOtherWrite) > 0 {
		dirPerms |= ModePermOtherExec
	}

	path := dir.path(name)
	if err := os.MkdirAll(filepath.Dir(path), dirPerms); err != nil {
		return err
	}

	return os.WriteFile(path, data, mode)
}

func (dir dirFS) Remove(name string) error {
	if err := checkname(name, "remove"); err != nil {
		return err
	}

	path := dir.path(name)
	// if file is not exist, implementation must throw an error.
	if _, err := os.Stat(path); err != nil {
		return err
	}

	return os.Remove(path)
}

func (dir dirFS) Symlink(oldname, newname string) error {
	if err := checkname(oldname, "symlink"); err != nil {
		return err
	}
	if err := checkname(newname, "symlink"); err != nil {
		return err
	}

	linkPath, err := filepath.Rel(filepath.Dir(dir.path(newname)), dir.path(oldname))
	if err != nil {
		fmt.Println(err)
		panic("unreachable")
	}

	return os.Symlink(linkPath, dir.path(newname))
}

func (dir dirFS) Readlink(name string) (string, error) {
	if err := checkname(name, "readlink"); err != nil {
		return "", err
	}

	path := dir.path(name)

	link, err := os.Readlink(path)
	if err != nil {
		return "", err
	}

	if filepath.IsAbs(link) {
		// TODO: IMPORTANT! We don't check here that symlink could potentially
		// go out of bounds.
		return strings.TrimPrefix(link, string(dir)+"/"), nil
	}

	return link, nil
}

func (dir dirFS) Lstat(name string) (FileInfo, error) {
	if err := checkname(name, "lstat"); err != nil {
		return nil, err
	}

	return os.Lstat(dir.path(name))
}

func checkname(name, op string) error {
	if !ValidPath(name) || runtime.GOOS == "windows" && strings.ContainsAny(name, `\:`) {
		return &PathError{Op: op, Path: name, Err: ErrInvalid}
	}

	return nil
}
