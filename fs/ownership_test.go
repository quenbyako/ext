// Copyright (c) 2020-2024 Richard Cooper
//
// This file is a part of quenbyako/ext package.
// See https://github.com/quenbyako/ext/blob/master/LICENSE for details

package fs_test

import (
	"io"
	"os"
	"os/user"
	"reflect"
	"runtime"
	"testing"
	"time"

	. "github.com/quenbyako/ext/fs"
)

func TestCheckPathAbleToWrite(t *testing.T) {
	t.Parallel()

	me := "100"
	myGroups := []string{"2"}

	exampleFS := testfs{
		"write/only/empty/dir": &testFile{finfo: &TestFileInfo{
			Perms: strPerms("rwxr-xr-x") | ModeDir,
			Uid:   "100",
			Gid:   "2",
		}},
		"write/only/empty/dir/root": &testFile{finfo: &TestFileInfo{
			Perms: strPerms("rwxr-xr-x") | ModeDir,
			Uid:   "0",
			Gid:   "0",
		}},
		"exec/only/file.txt": &testFile{finfo: &TestFileInfo{
			Perms: strPerms("--x--x--x"),
			Uid:   me,
			Gid:   "2",
		}},
	}

	for _, tt := range []struct {
		name    string
		fsys    FS
		path    string
		wantErr errorAssertionFunc
	}{{
		name: "exec perm only for existing file",
		fsys: exampleFS,
		path: "exec/only/file.txt",
		wantErr: exactErr(&PathError{Op: "write", Path: "exec/only/file.txt", Err: ErrPermissionExtended{
			Uid:  "100",
			Gid:  "2",
			Mode: strPerms("--x--x--x"),
		}}),
	}, {
		name: "no file in directory but can't create",
		fsys: exampleFS,
		path: "write/only/empty/dir/root/notexist.txt",
		wantErr: exactErr(&PathError{Op: "create", Path: "write/only/empty/dir/root", Err: ErrPermissionExtended{
			Uid:  "0",
			Gid:  "0",
			Mode: strPerms("rwxr-xr-x") | ModeDir,
		}}),
	}, {
		name: "no file in directory",
		fsys: exampleFS,
		path: "write/only/empty/dir/notexist.txt",
	}} {
		tt := tt
		tt.wantErr = noErrAsDefault(tt.wantErr)
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.wantErr(t, CheckFileAbleToWrite(tt.fsys, tt.path, me, myGroups))
		})
	}
}

func TestGetAllowedOperations(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name  string
		finfo FileInfo
		uid   string
		gids  []string
		want  Op
	}{} {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assertEqual(t, tt.want, GetAllowedOperations(tt.finfo, tt.uid, tt.gids))
		})
	}
}

func TestFileOwner(t *testing.T) {
	t.Parallel()

	// this is too simple test, just because git can't store uid and gid of
	// files, so it is impossible to generate test file on filesystem and check
	// different uids. So only one test.

	if !isUnix(runtime.GOOS) {
		// test makes no sense for non unix OSes
		return
	}

	f, err := os.CreateTemp("", "TestFileID.*.txt")
	noError(t, err)
	name := f.Name()
	u, err := user.Current()
	noError(t, err)

	defer os.Remove(name)

	// to be sure, that we stat'ted it
	f.Close()

	finfo, err := os.Stat(name)
	noError(t, err)

	gotUid, gotGid, gotOK := FileOwner(finfo)
	assertEqual(t, true, gotOK)
	assertEqual(t, u.Uid, gotUid)
	assertEqual(t, u.Gid, gotGid)

}

func noErrAsDefault(errfunc errorAssertionFunc) errorAssertionFunc {
	if errfunc == nil {
		return noError
	}
	return errfunc
}

// https://github.com/golang/go/blob/26b4844256/src/cmd/dist/build.go#L943
//
// No, this function is not exported 😭
func isUnix(s string) bool {
	switch s {
	case "aix", "android", "darwin", "dragonfly", "freebsd", "hurd", "illumos",
		"ios", "linux", "netbsd", "openbsd", "solaris":
		return true
	default:
		return false
	}
}

type testfs map[string]File

var _ FS = (testfs)(nil)

func (t testfs) Open(name string) (File, error) {
	if file, ok := t[name]; ok {
		if _, ok := file.(*forbiddenFile); ok {
			return nil, ErrPermission
		}

		return file, nil
	}

	return nil, ErrNotExist
}

type forbiddenFile struct{}

func (*forbiddenFile) Stat() (FileInfo, error)  { panic("unreachable") }
func (*forbiddenFile) Read([]byte) (int, error) { panic("unreachable") }
func (*forbiddenFile) Close() error             { panic("unreachable") }

type testFile struct {
	finfo FileInfo
}

func (t *testFile) Stat() (FileInfo, error)  { return t.finfo, nil }
func (_ *testFile) Read([]byte) (int, error) { return 0, io.EOF }
func (_ *testFile) Close() error             { return nil }

type TestFileInfo struct {
	Perms FileMode
	Uid   string
	Gid   string
}

var _ FileInfoOwner = (*TestFileInfo)(nil)

func (_ *TestFileInfo) Name() string             { return "undefined" }
func (_ *TestFileInfo) Size() int64              { return 0 }
func (i *TestFileInfo) Mode() FileMode           { return i.Perms }
func (_ *TestFileInfo) ModTime() time.Time       { return time.Time{} }
func (i *TestFileInfo) IsDir() bool              { return i.Perms.IsDir() }
func (_ *TestFileInfo) Sys() interface{}         { return nil }
func (i *TestFileInfo) Owner() (uid, gid string) { return i.Uid, i.Gid }

func strPerms(s string) (mode FileMode) {
	for i, m := range s {
		if m != '-' && i < 9 {
			mode |= (1 << uint(9-1-i))
		}
	}

	return mode
}

type errorAssertionFunc func(*testing.T, error)

func noError(t *testing.T, err error) {
	if err != nil {
		t.Log("expected no error")
	}
}

func exactErr(want error) errorAssertionFunc {
	return func(t *testing.T, err error) {
		assertEqual(t, want, err)
	}
}

func assertEqual[T any](t *testing.T, want, got T) {
	if !reflect.DeepEqual(want, got) {
		t.Logf("want: %#v, got %#v", want, got)
		t.FailNow()
	}
}
